# Porkbun DDNS

This project is a script to update Porkbun DNS to the IP of the current machine

This is used to provide some sort of dynamic DNS functionality when an internet provider does not provide a static IP

The script can be run via python directly, as a docker container, or as a helm chart

# Running the script

### via Python

The script can be run via python. It requires the environment variables API_KEY, SECRET_KEY, and DOMAIN, with TTL and SUBDOMAIN being optional parameters.

### via Docker

Docker build:
```shell
docker build . -t porkbun-ddns
```

Docker run:
```shell
docker run --rm -it \
  -e API_KEY="<PORKBUN API KEY>" \
  -e SECRET_KEY="<PORKBUN SECRET KEY>" \
  -e DOMAIN="<YOUR DOMAIN>" \
  -e SUBDOMAIN="<YOUR SUBDOMAIN>" \
  porkbun-ddns
```
Again, API_KEY, SECRET_KEY, and DOMAIN are required, with TTL and SUBDOMAIN being optional parameters.

### via Helm

Helm repository: oci://ghcr.io/willsunnn

Chart name: porkbun-dns-updater

Create secret:
```shell
kubectl create secret generic porkbun-api-creds \
  --from-literal=api_key='<YOUR_API_KEY>' \
  --from-literal=secret_key='<YOUR_SECRET_KEY>'
```

Example values: (see [values.yaml](helm/values.yaml) for defaults)
```yaml
fullnameOverride: porkbun-ddns
schedule: "*/1 * * * *"
env:
  DOMAIN: <domain to update>        # Required
  SUBDOMAIN: "*"
existingSecret:
  name: porkbun-api-creds           # Required
  apiKeyKey: api_key
  secretKeyKey: secret_key
```


# Creating a Docker and Helm Release

Set and push version tag
```shell
git tag 1.0.0
git push origin 1.0.0
```
[Build can be accessed here](https://github.com/willsunnn/porkbun-dns-updater/actions)
[Releases can be accessed here](https://github.com/willsunnn/porkbun-dns-updater/tags)
