/*
 * Temperature Checking Tool
 *
 * Description: A simple tool written in golang, for the purposes of
 *              monitoring temperaturatures of my devices in Linux.
 *              Specifically, this will work on kernel version 4.4+
 *
 * Author: Robert Bisewski <contact@ibiscybernetics.com>
 */

//
// Package
//
package main

//
// Imports
//
import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

//
// Globals
//
var (

	// Current location of the hardware sensor data, as of kernel 4.4+
	hardwareMonitorDirectory = "/sys/class/hwmon/"

	// Whether or not to print debug messages.
	debugMode = false

	// Attribute file for storing the hardware device name.
	hardwareNameFile = "name"

	// Attribute file for storing the hardware device current temperature.
	hardwareTempFile = "temp1_input"

	// Flag to check if the device in question uses the 'fam15h_power'
	// kernel module, in which case the temperature adjustment phase can be
	// skipped if the 'k10temp' module is also present.
	fam15hPowerModuleInUse = false

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

//! Function to handle printing debug messages when debug mode is on.
/*
 * @param      string    message to print to stdout
 *
 * @returns    none
 */
func debugPrint(debugMsg string) {

	// Return if debug mode is disabled.
	if debugMode != true {
		return
	}

	// Input validation.
	if len(debugMsg) < 1 {
		return
	}

	// Trim away unneeded whitespace.
	debugMsg = strings.Trim(debugMsg, " ")

	// Sanity check, make sure the pre-Trim'd string wasn't just whitespace.
	if len(debugMsg) < 1 {
		return
	}

	// Since this got a non-blank string, go ahead and print it to stdout.
	fmt.Println(debugMsg)
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
	fmt.Println("\n-----------------------------------------------")
	fmt.Println("Hardware Temperature Info Tool for Linux x86-64")
	fmt.Println("-----------------------------------------------\n")

	// Attempt to read in our file contents.
	listOfDeviceDirs, err := ioutil.ReadDir(hardwareMonitorDirectory)
	if err != nil {
		panic(err)
	}

	// Debug mode, print out a list of files in the directory specified by
	// the "hardwareMonitorDirectory" global variable.
	if debugMode {

		// Tell the end-user we are in debug mode.
		debugPrint("The following IDs are present in the hardware sensor " +
			"monitoring directory:\n")

		// String to hold out concat list of hardware device directories.
		debugStringForListOfDeviceDirs := ""

		// Cycle thru the array that holds the directories.
		for _, dir := range listOfDeviceDirs {
			debugStringForListOfDeviceDirs += "* " + dir.Name() + "\n"
		}

		// Finally, print out a list of device
		debugPrint(debugStringForListOfDeviceDirs)
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
		debugPrint(dir.Name() + " --> " +
			hardwareNameFilepathOfGivenDevice)

		// ...check to see if a 'name' file is present inside the directory.
		nameValueOfHardwareDevice, err := ioutil.ReadFile(
			hardwareNameFilepathOfGivenDevice)

		// If err is not nil, skip this device.
		if err != nil {

			// If debug mode, then print out a message telling the user
			// which device is missing a hardware 'name' file.
			debugPrint("Warning: " + dir.Name() + " does not contain a " +
				"hardware name file. Skipping...")

			// Move on to the next device.
			continue
		}

		// If the hardware name file does not contain anything of value,
		// skip it and move on to the next device.
		if len(nameValueOfHardwareDevice) < 1 {

			// If debug mode, then print out a message telling the user
			// which device is missing a hardware 'name' file.
			debugPrint("Warning: The hardware name file of " + dir.Name() +
				" does not contain valid data. Skipping...")

			// Move on to the next device.
			continue
		}

		// Trim away any excess whitespace from the hardware name file data.
		nameValueOfHardwareDeviceAsString :=
			strings.Trim(string(nameValueOfHardwareDevice), " \n")

		// Assemble the filepath to the temperature file of the currently
		// given hardware device.
		hardwareTemperatureFilepathOfGivenDevice :=
			hardwareMonitorDirectory + dir.Name() + "/" + hardwareTempFile

		// If debug mode, print out the current 'temperature' file we are
		// about to open.
		debugPrint(dir.Name() + " --> " +
			hardwareTemperatureFilepathOfGivenDevice)

		// If the hardware monitor contains an actively-updating temperature
		// sensor, then attempt to open it.
		temperatureValueOfHardwareDevice, err := ioutil.ReadFile(
			hardwareTemperatureFilepathOfGivenDevice)

		// If err is not nil, or temperature file is empty, tell the
		// end-user this device does not appear to have temperature data.
		if err != nil || len(temperatureValueOfHardwareDevice) < 1 {

			// If debug mode, then print out a message telling the user
			// which device is missing a hardware 'temperature' file.
			debugPrint("Warning: " + dir.Name() + " does not contain a " +
				"valid hardware temperature file, ergo no " +
				"temperature data to print for this device.")

			// append string values equivalent to the longest length.
			for len(nameValueOfHardwareDeviceAsString) <
				maxEntryLength+spacerSize {

				// append a space to the end of the string
				nameValueOfHardwareDeviceAsString += " "
			}

			// Print a none-available since no temperature data is available.
			fmt.Println(dir.Name(), " | ",
				nameValueOfHardwareDeviceAsString, "N/A\n")

			// With that done, go ahead and move on to the next device.
			continue
		}

		// If debug mode, tell the end-user that this is converted to
		// a string.
		debugPrint("Converting temperature file data from " +
			dir.Name() + " into a string.")

		// Attempt to convert the temperature to a string, trim it, and then
		// to an integer value afterwards.
		temperatureValueOfHardwareDeviceAsInt, err :=
			strconv.Atoi(strings.Trim(string(temperatureValueOfHardwareDevice), " \n"))

		// If err is not nil, then the temperature file does not have valid
		// integer data. So tell the end-user no data is available.
		if err != nil || len(temperatureValueOfHardwareDevice) < 1 {

			// If debug mode, then print out a message telling the user
			// which device is missing a hardware 'temperature' file.
			debugPrint("Warning: " + dir.Name() + " does not contain a " +
				"integer data in the hardware temperature file, " +
				"ergo no temperature data to print for this device.")

			// append string values equivalent to the longest length.
			for len(nameValueOfHardwareDeviceAsString) <
				maxEntryLength+spacerSize {

				// append a space to the end of the string
				nameValueOfHardwareDeviceAsString += " "
			}

			// Finally, print out the temperature data of the current device.
			fmt.Println(dir.Name(), " | ",
				nameValueOfHardwareDeviceAsString, "N/A\n")

			// With that done, go ahead and move on to the next device.
			continue
		}

		// Usually hardware sensors uses 3-sigma of precision and stores
		// the value as an integer for purposes of simplicity.
		//
		// Ergo, this needs to be divided by 1000 to give temperature
		// values that are meaningful to humans.
		//
		temperatureValueOfHardwareDeviceAsInt /= 1000

		// This acts as a work-around for the k10temp sensor module.
		if nameValueOfHardwareDeviceAsString == "k10temp" &&
			!fam15hPowerModuleInUse {

			// Add 30 degrees to the current temperature.
			temperatureValueOfHardwareDeviceAsInt += 30
		}

		// append string values equivalent to the longest length.
		for len(nameValueOfHardwareDeviceAsString) <
			maxEntryLength+spacerSize {

			// append a space to the end of the string
			nameValueOfHardwareDeviceAsString += " "
		}

		// Finally, print out the temperature data of the current device.
		fmt.Println(dir.Name(), " | ",
			nameValueOfHardwareDeviceAsString,
			temperatureValueOfHardwareDeviceAsInt, "C\n")
	}

	// If all is well, we can return quietly here.
	os.Exit(0)
}

// SetGlobalSensorFlags ... alters how Linux sees temperatures
/*
 * @param    os.FileInfo[]    array of directory data
 *
 * @return   error            error message, if any
 */
func SetGlobalSensorFlags(dirs []os.FileInfo) error {

	// input validation
	if dirs == nil || len(dirs) < 1 {
		return fmt.Errorf("SetGlobalSensorFlags() --> invalid input")
	}

	// Cycle thru the entire list of device directories...
	for _, dir := range dirs {

		// Assemble the filepath to the name file of the currently given
		// hardware device.
		hardwareNameFilepathOfGivenDevice := hardwareMonitorDirectory +
			dir.Name() + "/" + hardwareNameFile

		// If debug mode, print out the current 'name' file we are about
		// to open.
		debugPrint(dir.Name() + " --> " +
			hardwareNameFilepathOfGivenDevice)

		// ...check to see if a 'name' file is present inside the directory.
		nameValueOfHardwareDevice, err := ioutil.ReadFile(
			hardwareNameFilepathOfGivenDevice)

		// If err is not nil, skip this device.
		if err != nil {

			// If debug mode, then print out a message telling the user
			// which device is missing a hardware 'name' file.
			debugPrint("Warning: " + dir.Name() + " does not contain a " +
				"hardware name file. Skipping...")

			// Move on to the next device.
			continue
		}

		// If the hardware name file does not contain anything of value,
		// skip it and move on to the next device.
		if len(nameValueOfHardwareDevice) < 1 {

			// If debug mode, then print out a message telling the user
			// which device is missing a hardware 'name' file.
			debugPrint("Warning: The hardware name file of " + dir.Name() +
				" does not contain valid data. Skipping...")

			// Move on to the next device.
			continue
		}

		// Trim away any excess whitespace from the hardware name file data.
		nameValueOfHardwareDeviceAsString :=
			strings.Trim(string(nameValueOfHardwareDevice), " \n")

		// Determine the length of the longest entry string
		//
		// TODO: this is a less than ideal place for this code, consider
		//       rewriting how this program handles hwmonX entries at some
		//       future date
		//
		if len(nameValueOfHardwareDeviceAsString) > maxEntryLength {
			maxEntryLength = len(nameValueOfHardwareDeviceAsString)
		}

		// Conduct a quick check to determine if the 'fam15h_power' module
		// is currently in use.
		if nameValueOfHardwareDeviceAsString == "fam15h_power" {
			fam15hPowerModuleInUse = true
		}
	}

	// everything worked fine, so return null
	return nil
}
