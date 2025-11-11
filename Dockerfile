FROM python:3.14-alpine

RUN addgroup -S ddns && adduser -S ddns -G ddns

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY porkbun_ddns.py .
RUN chmod +x porkbun_ddns.py

RUN mkdir /cache && chown ddns:ddns /cache

USER ddns

CMD ["python", "/app/porkbun_ddns.py"]
