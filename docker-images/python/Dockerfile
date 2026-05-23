FROM python:3.11-slim

WORKDIR /app

RUN useradd -m sandboxuser

USER sandboxuser

CMD ["python", "/app/main.py"]