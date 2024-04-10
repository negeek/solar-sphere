# Solar-Sphere

Solar-Sphere is a free and powerful API designed to provide remote access to solar irradiance readings.

## The Device

![alt text](https://github.com/negeek/solar-sphere/blob/main/solarmeterproject.png)

### About the API

Solar-Sphere offers a Low-Cost Solar-Irradiance meter designed to achieve high accuracy similar to thermopile-based pyranometers. The API is built to support multiple Solar-Irradiance devices, each with its unique ID.

### Authentication

To access the API, users need to authenticate using a Device ID. Here's an example of how to authenticate:

`curl -X POST 'http://localhost:8080/auth/v1/join/' -H 'Content-Type: application/json' -d'{"email":"patrick@gmail.com", "device_id":"75860507183551752178871-"}'`

### Solar Irradiance Data Collection
You can publish solar irradiance data to the specified MQTT topic and MQTT broker:

Broker URL: `tcp://broker.emqx.io:1883`

Topic format: `solar-sphere/solar-sentinel/sensor/<device_id>/solar-irradiance`

### Solar Irradiance Data Download
`curl -X GET 'http://localhost:8080/sentinel/v1/download/{device_id}' -H 'Content-Type: application/json' -H 'Authorisation: Bearer eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJFbWFpbCI6ImFkYW5pbmlwYXRyaWNrMTdAZ21haWwuY29tIiwiRGF0ZVRpbWUiOiIyMDI0LTA0LTEwVDIyOjM3OjI3LjE4Mjk4ODE2M1oifQ.hT1jaG92Kcpj1rvPVjq4i-YB21yq8ugRB4suUYOQhCN5IpFhDkSqcbU4nmDuOieCjedtb7pLvaBKfyzvw01xAg'`


###### Note: Project still in development
