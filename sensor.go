package main

type Sensor struct {

        // name of sensor
        name string

        // location to the OS path
        path string

        // sensor type; e.g. temp for Temperature sensors or fan for Fan sensors
        category string

        // refined sensor data, as an int
        intData int

        // current sensor number, for a given category, for a given hwmon; e.g. temp sensor 3 of a device with 5 temp sensors
        number int

        // maximum number of sensors, for a given category, for a given hwmon
        count int
}
