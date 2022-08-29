FROM python:3.10-slim

WORKDIR /src

COPY ./requirements.txt .

RUN pip install --no-cache-dir --upgrade -r requirements.txt

COPY . /src

EXPOSE 8080

CMD ["uvicorn", "main:app", "--port=8080", "--host=0.0.0.0"]