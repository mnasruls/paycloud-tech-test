# paycloud-tech-test

If you want to run this test, please follow these steps:

1. Clone this repository.
2. Change directory to test you want to run.
3. Use this command `go run .` instead of using `go run main.go`.

If your don't have rabbitmq in your local. You can use this docker compose

````
version: '3'

services:
    rabbit-1:
        image: rabbitmq:3.10.5-management-alpine
        hostname: rabbit-1
        container_name: rabbit-1
        ports:
            - '8080:15672'
            - '8081:5672'
        environment:
            - RABBITMQ_DEFAULT_USER=guest
            - RABBITMQ_DEFAULT_PASS=guest
        networks:
            - gomc-broker
networks:
    gomc-broker:
    ```
````
