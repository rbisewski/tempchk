package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"strconv"
)

//! Function to handle printing debug messages when debug mode is on.
/*
 * @param      string    message to print to stdout
 *
 * @returns    none
 */
func debug(debugMsg string) {

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

//! Obtains hwmon sensor data.
/*
 * @param      string    name of device
 * @param      string    full path of the given hwmon directory
 *
 * @returns    error     whether or not the output is feasible
 *             int       sensor data, as an integer; e.g. degrees C or RPM
 */
func GetSensorData(name string, hwmon string) (error, int) {

        // input validation
        if name == "" || hwmon == "" {
                return fmt.Errorf("GetSensorData(): invalid input"), -1
        }

	// Assemble the filepath to the temperature file of the currently
	// given hardware device.
	path := hardwareMonitorDirectory + hwmon + "/" +
                tempPrefix + "1" + inputSuffix

	// If debug mode, print out the current 'temperature' file we are
	// about to open.
	debug(hwmon + " --> " + path)

	// If the hardware monitor contains an actively-updating temperature
	// sensor, then attempt to open it.
	rawData, err := ioutil.ReadFile(path)

	// If err is not nil, or temperature file is empty, tell the
	// end-user this device does not appear to have temperature data.
	if err != nil || len(rawData) < 1 {
                return err, -1
	}

	// If debug mode, tell the end-user that this is converted to
	// a string.
	debug("Converting temperature file data from " +
		hwmon + " into a string.")

	// Attempt to convert the temperature to a string, trim it, and then
	// to an integer value afterwards.
	trimmedIntData, err := strconv.Atoi(strings.Trim(string(rawData), " \n"))

        return err, trimmedIntData
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
		return fmt.Errorf("SetGlobalSensorFlags(): invalid input")
	}

	// Cycle thru the entire list of device directories...
	for _, dir := range dirs {

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
			digitalAmdPowerModuleInUse = true
		}
	}

	//
	// attempt to read from the CPU info file to determine if Ryzen
	//

	cpuinfoFileAsBytes, err := ioutil.ReadFile(cpuinfoDirectory)
	if len(cpuinfoFileAsBytes) == 0 || err != nil {
		return nil
	}

	cpuinfoString := string(cpuinfoFileAsBytes)
	if cpuinfoString == "" {
		return nil
	}

	if strings.Contains(cpuinfoString, "Ryzen") {
		digitalAmdPowerModuleInUse = true
	}

	// everything worked fine, so return null
	return nil
}
