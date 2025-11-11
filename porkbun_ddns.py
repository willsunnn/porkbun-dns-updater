#!/usr/bin/env python3

import os
import sys
from typing import Dict
import requests
import logging

def post(
        path: str,
        api_key: str,
        secret_key: str,
        body: Dict[str,str]=dict(),
    ) -> requests.Response:
    base_url = "https://api.porkbun.com/api/json/v3/"
    url = base_url + path
    payload = body | {
        "apikey": api_key,
        "secretapikey": secret_key,
    }
    return requests.post(url, json=payload)

def get_ip(api_key: str, secret_key: str):
    """Fetches current IP as seen from porkbun servers"""
    response = post("ping", api_key, secret_key)
    if response.status_code != 200:
        raise RuntimeError(f"Failed to get current IP status={response.status_code} response={response.text}")
    return response.json()["yourIp"]

def get_dns_records(api_key: str, secret_key: str, domain: str, subdomain: str | None=None):
    """Fetches the current A record from Porkbun."""
    subpath = f"dns/retrieveByNameType/{domain}/A/{subdomain}" if (subdomain is not None) else f"dns/retrieveByNameType/{domain}/A/"
    response = post(subpath, api_key, secret_key)
    if response.status_code != 200:
        raise RuntimeError(f"Failed to fetch existing DNS record status={response.status_code} response={response.text}")
    body = response.json()
    if body["status"] != "SUCCESS":
        raise RuntimeError(f"Failed to fetch existing DNS record status={response.status_code} response={response.text}")
    return body["records"]

def update_dns_record(api_key: str, secret_key: str, ip: str, domain: str, subdomain: str | None=None, ttl: int=600):
    """Updates the A record in Porkbun."""
    subpath = f"dns/editByNameType/{domain}/A/{subdomain}" if (subdomain is not None) else f"dns/editByNameType/{domain}/A/"
    response = post(subpath, api_key, secret_key, {
        "content": ip,
        "TTL": ttl,
    })
    if response.status_code != 200 or response.json()["status"] != "SUCCESS":
        raise RuntimeError(f"Failed to update existing DNS record status={response.status_code} response={response.text}")

def create_dns_record(api_key: str, secret_key: str, ip: str, domain: str, subdomain: str | None=None, ttl: int=600) -> bool:
    """Creates the A record in Porkbun."""
    subpath = f"dns/create/{domain}"
    response = post(subpath, api_key, secret_key, {
        "type": "A",
        "name": subdomain,
        "content": ip,
        "TTL": ttl,
    })
    if response.status_code != 200 or response.json()["status"] != "SUCCESS":
        raise RuntimeError(f"Failed to create DNS record status={response.status_code} response={response.text}")

def upsert_dns_record(api_key: str, secret_key: str, ip: str, domain: str, subdomain: str | None=None, ttl: int=600):
    """Updates or Creates the A record in Porkbun."""
    records = get_dns_records(api_key, secret_key, domain, subdomain)
    if len(records) > 1:
        raise RuntimeError(f"Could not update DNS record as multiple records were found:\n{records}")
    elif len(records) == 0:
        create_dns_record(api_key, secret_key, ip, domain, subdomain, ttl)
        logging.info(f"Created dns record with ip {ip}")
    elif records[0]["content"] == ip:
        logging.info(f"Skipping update as record already matches current ip {ip}")
    else:
        update_dns_record(api_key, secret_key, ip, domain, subdomain, ttl)
        logging.info(f"Updated dns record with ip {ip}")
        


def main():
    # Configure logger
    root = logging.getLogger()
    root.setLevel(logging.INFO)
    handler = logging.StreamHandler(sys.stdout)
    handler.setLevel(logging.INFO)
    formatter = logging.Formatter('%(asctime)s - %(levelname)s - %(message)s')
    handler.setFormatter(formatter)
    root.addHandler(handler)

    # Parse env variables
    api_key = os.environ.get("API_KEY")
    secret_key = os.environ.get("SECRET_KEY")
    domain = os.environ.get("DOMAIN")
    subdomain = os.environ.get("SUBDOMAIN", None)   # Optional
    ttl = int(os.environ.get("TTL", "600"))         # Optional

    if not all([api_key, secret_key, domain]):
        logging.error("API_KEY, SECRET_KEY, and DOMAIN are required")
        sys.exit(1)
    
    full_domain = f"{subdomain}.{domain}" if subdomain else domain
    logging.info(f"Updating porkbun for {full_domain}")

    try:
        ip = get_ip(api_key, secret_key)
        upsert_dns_record(api_key, secret_key, ip, domain, subdomain, ttl)
        logging.info(f"DNS update/check complete. ip={ip}")
    except Exception as e:
        logging.exception(f"An unexpected error occurred: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()