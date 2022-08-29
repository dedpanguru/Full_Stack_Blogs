import os

from bson.errors import InvalidDocument
from fastapi import FastAPI, HTTPException
from fastapi.encoders import jsonable_encoder
import motor.motor_asyncio
from pymongo.errors import DuplicateKeyError
from pymongo.results import UpdateResult
from starlette import status
from starlette.responses import JSONResponse

from models import Post, DateOfCreation

app = FastAPI()

collection = motor.motor_asyncio.AsyncIOMotorClient(os.environ["DB_ADDRESS"]).blog.posts
collection.create_index([('year', 1), ('month', 1), ('day', 1)], unique=True)


@app.get("/posts/{year}/{month}/{day}")
async def get_posts(year: int, month: int, day: int):
    # get all posts, sorted by date
    filter = {}
    if year:
        filter["year"] = {"$eq": year}
    if month:
        filter["month"] = {"$eq": month}
    if day:
        filter["day"] = {"$eq": day}
    posts = await collection.find(filter).sort(["year", "month", "day"]).to_list(10)
    if not posts:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail={"message":f"post with date {year}-{month}-{day} no found"})
    return JSONResponse(status_code=status.HTTP_200_OK, content=posts)


@app.get('/posts')
async def get_all_posts():
    posts = await collection.find({}).to_list(10)
    return JSONResponse(status_code=status.HTTP_200_OK, content=posts)


@app.post("/new", response_description="Creates a new post", response_model=Post, status_code=status.HTTP_201_CREATED)
async def create_post(post: Post):
    post_as_dict = jsonable_encoder(post)
    try:
        await collection.insert_one(post_as_dict)
    except DuplicateKeyError:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=f'A post with date {post.year}-{post.month}-{post.day} already exists!')


@app.put("/edit", response_description='update post', response_model=Post, status_code=status.HTTP_202_ACCEPTED)
async def update_post(post: Post):
    post_as_dict = jsonable_encoder(post)
    try:
        result: UpdateResult = await collection.update_one(post.get_date_as_map(), {'$set': post_as_dict})
        if not result.:
            raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=f'Post with date {post.year}-{post.month}-{post.day} not found!')
    except InvalidDocument:
        raise HTTPException(status_code=status.HTTP_422_UNPROCESSABLE_ENTITY, detail='invalid request body')


@app.delete("/delete", response_description='Deletes a post', response_model=DateOfCreation, status_code=status.HTTP_202_ACCEPTED)
async def delete_post(d: DateOfCreation):
    result = await collection.delete_one({"year": d.year, "month": d.month, "day": d.day})
    if result.deleted_count != 1:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=f"Post with date {d.year}-{d.month}-{d.day} not found!")


@app.get("/hello/{name}")
async def say_hello(name: str):
    return {"message": f"Hello {name}"}
