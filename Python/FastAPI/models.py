from datetime import datetime, date
from typing import Optional

from pydantic import BaseModel, Field
from bson import ObjectId


class PyObjectId(ObjectId):
    @classmethod
    def __get_validators__(cls):
        yield cls.validate

    @classmethod
    def validate(cls, v):
        if not ObjectId.is_valid(v):
            raise ValueError("Invalid objectid")
        return ObjectId(v)

    @classmethod
    def __modify_schema__(cls, field_schema):
        field_schema.update(type="string")


class Post(BaseModel):
    id: PyObjectId = Field(default_factory=PyObjectId, alias="_id")
    title: str = Field(...)
    content: str = Field(...)
    created_at: Optional[str] = Field(default=datetime.now().astimezone().replace(microsecond=0).isoformat(), alias="createdAt")
    updated_at: Optional[str] = Field(default=datetime.now().astimezone().replace(microsecond=0).isoformat(), alias="updatedAt")
    year: int = Field(default=date.today().year)
    month: int = Field(default=date.today().month)
    day: int = Field(default=date.today().day)

    class Config:
        allow_population_by_field_name = True
        arbitrary_types_allowed = True
        json_encoders = {ObjectId: str}
        schema_extra = {
            "example": {
                "title": "test",
                "year": 2022,
                "month": 11,
                "day": 23,
                "createdAt":"2020-03-20T14:31:43+13:00",
                "updatedAt":"2020-03-20T14:31:43+13:00",
                "content": "test"
            }
        }

    def change_updated_at(self):
        self.updated_at = datetime.now().astimezone().replace(microsecond=0).isoformat()

    def get_date_as_map(self) -> dict:
        return {"year": self.year, "month": self.month, "day": self.day}

class DateOfCreation(BaseModel):
    year: int = Field(default=date.today().year)
    month: int = Field(default=date.today().month)
    day: int = Field(default=date.today().day)