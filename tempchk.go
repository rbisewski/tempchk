package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

//
// Globals
//
var (
	// cpu info location, as of kernel 4.4+
	cpuinfoDirectory = "/proc/cpuinfo"

	// Current location of the hardware sensor data, as of kernel 4.4+
	hardwareMonitorDirectory = "/sys/class/hwmon/"

	// Whether or not to print debug messages.
	debugMode = false

	// Attribute file for storing the hardware device name.
	hardwareNameFile = "name"

	// Attribute file for storing the hardware device current temperature.
        tempPrefix = "temp"
        inputSuffix = "_input"

	// flag to check whether the AMD digital thermo module is in use
	digitalAmdPowerModuleInUse = false

	// size of the longest hwmonX/name entry string
	maxEntryLength = 0

	// spacer size
	spacerSize = 4

	// Whether or not to print the current version of the program
	printVersion = false

	// default version value
	Version = "0.0"
)

// Initialize the argument input flags.
func init() {

	// Version mode flag
	flag.BoolVar(&printVersion, "version", false,
		"Print the current version of this program and exit.")
}

//
// PROGRAM MAIN
//
func main() {

	// Parse the flags, if any.
	flag.Parse()

	// if requested, go ahead and print the version; afterwards exit the
	// program, since this is all done
	if printVersion {
		fmt.Println("tempchk v" + Version)
		os.Exit(0)
	}

	// Print out a few lines telling the user that the program has started.
	fmt.Println("\nCurrent temperature sensor readings\n---")

	// Attempt to read in our file contents.
	listOfDeviceDirs, err := ioutil.ReadDir(hardwareMonitorDirectory)
	if err != nil {
		panic(err)
	}

	// Debug mode, print out a list of files in the directory specified by
	// the "hardwareMonitorDirectory" global variable.
	if debugMode {

		// Tell the end-user we are in debug mode.
		debug("The following IDs are present in the hardware sensor " +
			"monitoring directory:\n")

		// String to hold out concat list of hardware device directories.
		debugStringForListOfDeviceDirs := ""

		// Cycle thru the array that holds the directories.
		for _, dir := range listOfDeviceDirs {
			debugStringForListOfDeviceDirs += "* " + dir.Name() + "\n"
		}

		// Finally, print out a list of device
		debug(debugStringForListOfDeviceDirs)
	}

	// Search thru the directories and set the relevant flags...
	err = SetGlobalSensorFlags(listOfDeviceDirs)

	// safety check, ensure no errors occurred
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// For each of the devices...
	for _, dir := range listOfDeviceDirs {

		// Assemble the filepath to the name file of the currently given
		// hardware device.
		hardwareNameFilepathOfGivenDevice := hardwareMonitorDirectory +
			dir.Name() + "/" + hardwareNameFile

		// If debug mode, print out the current 'name' file we are about
		// to open.
		debug(dir.Name() + " --> " +
			hardwareNameFilepathOfGivenDevice)

		// ...check to see if a 'name' file is present inside the directory.
		nameValueOfHardwareDevice, err := ioutil.ReadFile(
			hardwareNameFilepathOfGivenDevice)

		// If err is not nil, skip this device.
		if err != nil {

			// If debug mode, then print out a message telling the user
			// which device is missing a hardware 'name' file.
			debug("Warning: " + dir.Name() + " does not contain a " +
				"hardware name file. Skipping...")

			// Move on to the next device.
			continue
		}

		// If the hardware name file does not contain anything of value,
		// skip it and move on to the next device.
		if len(nameValueOfHardwareDevice) < 1 {

			// If debug mode, then print out a message telling the user
			// which device is missing a hardware 'name' file.
			debug("Warning: The hardware name file of " + dir.Name() +
				" does not contain valid data. Skipping...")

			// Move on to the next device.
			continue
		}

		// Trim away any excess whitespace from the hardware name file data.
		trimmedName := strings.Trim(string(nameValueOfHardwareDevice), " \n")

                sensors, err := GetSensorData(trimmedName, dir.Name())

		// If err is not nil, then the temperature file does not have valid
		// integer data. So tell the end-user no data is available.
		if err != nil || len(sensors) < 1 {

			// If debug mode, then print out a message telling the user
			// which device is missing a hardware 'temperature' file.
			debug("Warning: " + dir.Name() + " does not contain " +
				"valid sensor data in the hardware input file, " +
				"ergo no temperature data to print for this device.")

			// append string values equivalent to the longest length.
			for len(trimmedName) < maxEntryLength+spacerSize {
				trimmedName += " "
			}

			// Finally, print out the temperature data of the current device.
			fmt.Println(dir.Name(), "  ",
				trimmedName, "N/A")

			// With that done, go ahead and move on to the next device.
			continue
		}

                for _, sensor := range sensors {

                        // Usually hardware sensors uses 3-sigma of precision and stores
                        // the value as an integer for purposes of simplicity.
                        //
                        // Ergo, this needs to be divided by 1000 to give temperature
                        // values that are meaningful to humans.
                        //
                        sensor.intData /= 1000

                        // This acts as a work-around for the k10temp sensor module.
                        if trimmedName == "k10temp" &&
				!digitalAmdPowerModuleInUse {

				// Add 30 degrees to the current temperature.
				sensor.intData += 30
                        }

                        // append string values equivalent to the longest length.
                        for len(trimmedName) <
				maxEntryLength+spacerSize {

				// append a space to the end of the string
				trimmedName += " "
                        }

                        // Finally, print out the temperature data of the current device.
                        fmt.Println(dir.Name(), "  ",
				trimmedName,
				sensor.intData, "C")
                }
	}

	// If all is well, we can return quietly here.
	os.Exit(0)
}
