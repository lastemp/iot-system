# iot-system

This is an IOT system that comprises of two sub systems as indicated below;

## edge-client
This is the client deployed on the edge that connects to mqtt broker and subscribes on the given topics.
On receiving the mqtt messages from the broker, it queues the messages for a given duration before posting them to the cloud api url.

## cloud-restful-api
This is a RESTful Gin Web API that connects to MySQL database. It receives posted messages from edge-client and then inserts them to database.

Currently this RESTful API supports: 
- Register messages
- Register batch messages

You'll need to have a MySQL (or compatible) server running on your machine to test this example.

## Usage

All the following commands assume that your current working directory is _this_ directory. I.e.:

```console
$ pwd
.../iot-system
```

1. Create database and tables:

   The `sql` directory contains the SQL files used for database setup:
   
   Database
   ```sh
   mysql -u root -p < sql/0_create_database.sql
   ```
   
   Tables
   ```sh
   mysql -u root -p iot-system < sql/tables/*.sql
   ```

   For each step you will be prompted for the root user's password. If there's no password set on the root use, just hit enter again.

1. Create a `.env` file in the directory [edge-client](./edge-client/) :

   ```ini
   MQTT_BROKER_ADDR=tcp://test.mosquitto.org:1883
   CLIENT_ID=test-mqtt-client
   TOPIC=sensors/#
   BATCHMESSAGE_API_URL=http://localhost:8080/batchmessage
   ```
   
1. Create a `.env` file in the directory [cloud-restful-api](./cloud-restful-api/) :

   ```ini
   SERVER_ADDR=127.0.0.1:8080
   DB_USER=root
   DB_PASSWORD=<password>
   DB_HOST_PORT=127.0.0.1:3306
   DB_NAME=ebusiness_iot
   ```

   Update "DB_USER" and "DB_PASSWORD" values with the correct MySQL user/password.

1. Run the application in directory [edge-client](./edge-client/) :

   ```sh
   go run .
   ```

1. Run the application in directory [cloud-restful-api](./cloud-restful-api/) :

   ```sh
   go run .
   ```
