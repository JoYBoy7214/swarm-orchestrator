import asyncio
import signal
import json
import logging

from nats.aio.client import Client as NATS
from nats.errors import TimeoutError
from qdrant_client.models import PointStruct

# Import your custom database class from the other file
from db_store import QdrantStore

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

async def worker_1(js, stop_event, store,subject_string):
    try:
        sub = await js.pull_subscribe(subject_string, "py_scraper_worker", stream="myswarm")
       
    except Exception as e:
        print(f"Error creating consumer: {e}")
        raise

    while not stop_event.is_set():
        try:
            # Fetch 1 message, wait up to 1 second
            msgs = await sub.fetch(1, timeout=1.0)
            
            for msg in msgs:
                d = json.loads(msg.data)
                
                # Extract values from your JSON payload 
                msg_id = d.get("event_id","")
                msg_task_type=d.get("user_req",{}).get("task_type","")
                msg_url=d.get("user_req",{}).get("payload",{}).get("url","")
                msg_status=d.get("status","")
                msg_created=d.get("created_at","")
                msg_vector=[0.05, 0.61, 0.76, 0.74] #just for testing
                meta_data={
                    "url":msg_url,
                    "task_type": msg_task_type,
                    "status":msg_status,
                    "created_at":msg_created,
                }
                if(msg_id==""):
                    print("error task without UUID")
                    await msg.ack()
                    continue
                my_data = [
                    PointStruct(
                        id=msg_id,
                        vector=msg_vector,
                        payload=meta_data
                    )
                ]
                
                try:  
                # AWAIT the upsert function!
                    await store.upsert_points("test", my_data)
                    await asyncio.sleep(3)
                    print(f"Processed message: {d}")
                    
                except Exception as e:
                    logger.error(f"Database processing failed for message {msg.sequence}: {e}")
                    try:
                        await msg.nak()
                    except Exception:
                        pass
                    continue
                try:
                    pub_payload = {"event_id": msg_id, "status": "completed"}
                    json_payload = json.dumps(pub_payload).encode("utf-8")
                    await js.publish("task.status.completed", json_payload)
                    await msg.ack()
                except Exception as pub_err:
                    await msg.nak()
                    print(f"Task saved, but failed to publish completion: {pub_err}")
                
        except TimeoutError:
            continue  # No messages, loop again and check stop_event
        except asyncio.CancelledError: #ctrl will trigger t1.cancel by event loop or main thread 
            # Handle task cancellation smoothly on shutdown
            break
        except Exception as e:
            if not stop_event.is_set():
                print(f"Worker 1 error: {e}")
                


async def main():
    # 1. Initialize the Database Store using the imported class
    store = QdrantStore()
    await store.create_new_collection("test")
    # 2. Connect to NATS
    nc = NATS()
    try:
        await nc.connect("nats://nats:4222", connect_timeout=30)
    except Exception as e:
        print(f"Error connecting to NATS: {e}")
        return

    # 3. Setup JetStream Context
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

    # 4. Start Consuming
   # t1 = asyncio.create_task(worker_1(js, stop_event, store,"task.scrape.pdf")) #This will schelude the coroutine not run it 
   
    try:
        while not stop_event.is_set():
            await asyncio.sleep(0.5)
    except KeyboardInterrupt:
        stop_event.set()

    # 5. Graceful Shutdown
    # try:
    #     await t1
    # except asyncio.CancelledError:
    #     # This is expected behavior when calling task.cancel()
    #     pass
    # except Exception as e:
    #     print(f"Worker exited with error during shutdown: {e}")
    

    await nc.close()

if __name__ == '__main__':
    asyncio.run(main())