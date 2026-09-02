import asyncio
import signal
import json
import logging
from nats.aio.client import Client as NATS
from nats.errors import TimeoutError

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

async def worker_task(worker_id, js, stop_event):
    try:
        # All concurrent workers share this durable name.
        # JetStream automatically load-balances the messages across them.
        sub = await js.pull_subscribe("task.EXECUTE", durable="concurrent_worker_group")
    except Exception as e:
        logger.error(f"Worker {worker_id} - Error creating consumer: {e}")
        return

    logger.info(f"Worker {worker_id} started.")

    while not stop_event.is_set():
        try:
            # Fetch 1 message, wait up to 1 second
            msgs = await sub.fetch(1, timeout=0.5)
            
            for msg in msgs:
                try:
                    # 1. Parse the schema
                    d = json.loads(msg.data)
                    workflow_id = d.get("Workflow_id")
                    task_id = d.get("Task_id")
                    task_type = d.get("Task_type")
                    
                    logger.info(f"Worker {worker_id} processing Task: {task_id} (Type: {task_type})")

                    # 2. Simulate task execution
                    await asyncio.sleep(2)
                    
                    # 3. Prepare completed payload
                    completed_payload = {
                        "Workflow_id": workflow_id,
                        "Task_id": task_id,
                        "Task_type": task_type,
                    }
                    
                    # 4. Push to completion subject
                    await js.publish(
                        "task.COMPLETED", 
                        json.dumps(completed_payload).encode("utf-8")
                    )
                    
                    # 5. Acknowledge successful processing
                    await msg.ack()
                    logger.info(f"Worker {worker_id} completed and acked Task: {task_id}")
                    
                except json.JSONDecodeError:
                    logger.error(f"Worker {worker_id} - Invalid JSON. Terminating message.")
                    # .term() prevents the message from being requeued
                    await msg.term()
                    
                except Exception as e:
                    logger.error(f"Worker {worker_id} - Processing failed: {e}")
                    # .nak() tells JetStream we failed and to redeliver it
                    await msg.nak()
                
        except TimeoutError:
            continue  # No messages, loop again and check stop_event
        except asyncio.CancelledError:
            break
        except Exception as e:
            if not stop_event.is_set():
                logger.error(f"Worker {worker_id} unexpected error: {e}")
                await asyncio.sleep(1) # Prevent tight loop on severe connection issues
                
    logger.info(f"Worker {worker_id} shutting down.")


async def main():
    nc = NATS()
    try:
        await nc.connect("localhost:4222", connect_timeout=30)
    except Exception as e:
        logger.error(f"Error connecting to NATS: {e}")
        return

    js = nc.jetstream()
    stop_event = asyncio.Event()

    def signal_handler():
        stop_event.set()

    loop = asyncio.get_running_loop()
    for sig in (signal.SIGINT, signal.SIGTERM):
        try:
            loop.add_signal_handler(sig, signal_handler)
        except NotImplementedError:
            pass 

    # --- Start Concurrent Workers ---
    NUM_WORKERS = 5
    workers = []
    
    for i in range(NUM_WORKERS):
        task = asyncio.create_task(worker_task(i + 1, js, stop_event))  #this will schedule corountines in the background but it won't run it 
        workers.append(task)
   
    try:
        while not stop_event.is_set():
            await asyncio.sleep(0.5)  #this juggle between main func and corountines 
    except KeyboardInterrupt:
        stop_event.set()

    # --- Graceful Shutdown ---
    logger.info("Initiating graceful shutdown...")
    
    # Wait for all workers to finish their current loop iteration
    await asyncio.gather(*workers, return_exceptions=True)
    await nc.close()
    logger.info("Shutdown complete.")

if __name__ == '__main__':
    asyncio.run(main())