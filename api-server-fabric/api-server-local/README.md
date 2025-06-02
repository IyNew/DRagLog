# Local API Documentation

## Overview

This project provides a local RESTful API service written in Go for certificate management, using CouchDB as the backend database. The service is containerized and managed via Docker Compose for easy local development and testing.

---

## Start the Service

Make sure you have [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/) installed.

From the `localapi` directory, run:

```bash
docker compose up --build
```
This command will build the Docker images and start the containers defined in the `docker-compose.yml` file. 

The API will be available at `http://localhost:8080`, you can get a visulized view of API at `http://localhost:8080/docs#`

The CouchDB service will be at `http://localhost:5984`, you can access the CouchDB web interface at `http://localhost:5984/_utils/` (username: `admin`, password: `admin123`).



