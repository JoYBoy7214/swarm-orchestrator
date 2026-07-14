import logging
from qdrant_client import AsyncQdrantClient
from qdrant_client.models import VectorParams, Distance, PointStruct

class QdrantStore:
    def __init__(self):
        try:
            # Use AsyncQdrantClient to prevent blocking the event loop
            self.client = AsyncQdrantClient(host="qdrant", grpc_port=6334, prefer_grpc=True)
            print("Successfully initialized Async Qdrant Client!")
        except Exception as e:
            logging.error(f"Failed to create client: {e}")
            raise e
    
    async def create_new_collection(self, collection_name: str):
        try:
            exists = await self.client.collection_exists(collection_name)
            if not exists:
                await self.client.create_collection(
                    collection_name=collection_name,
                    vectors_config=VectorParams(
                        size=4, 
                        distance=Distance.COSINE 
                    ),
                )  
                print(f"Collection '{collection_name}' created.")
            else:
                print(f"Collection '{collection_name}' already exists.") 
        except Exception as e:
            logging.error(f"Failed to create collection: {e}")
            raise e
        
    async def upsert_points(self, collection_name: str, points: list[PointStruct]):
        try:
            operation_info = await self.client.upsert(
                collection_name=collection_name,
                points=points
            )
            print(f"Successfully upserted {len(points)} points. Status: {operation_info.status}")
        except Exception as e:
            logging.error(f"Failed to upsert points: {e}")
            raise e