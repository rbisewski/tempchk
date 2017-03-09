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
    hardware_monitor_directory = "/sys/class/hwmon/"

    // Whether or not to print debug messages.
    debug_mode = false

    // Attribute file for storing the hardware device name.
    hardware_name_file = "name"

    // Attribute file for storing the hardware device current temperature.
    hardware_temp_file = "temp1_input"
)

//! Function to handle printing debug messages when debug mode is on.
/*
 * @param      string    message to print to stdout
 *
 * @returns    none
 */
func debug_print(debug_msg string) {

    // Return if debug mode is disabled.
    if debug_mode != true {
        return
    }

    // Input validation.
    if len(debug_msg) < 1 {
        return
    }

    // Trim away unneeded whitespace.
    debug_msg = strings.Trim(debug_msg," ")

    // Sanity check, make sure the pre-Trim'd string wasn't just whitespace.
    if len(debug_msg) < 1 {
        return
    }

    // Since this got a non-blank string, go ahead and print it to stdout.
    fmt.Println(debug_msg)
}

//! Function to panic if an error has occurred.
/*
 * @param      error    given golang error value
 * 
 * @returns    none
 */
func panic_if_error(e error) {

    // Input validation.
    if e == nil {
        return
    }

    // Tell the end user something bad has occurred, and then do a panic()
    fmt.Println("Error: The following critical issue has occurred...")
    panic(e)
}


//
// PROGRAM MAIN
//
func main() {

    //
    // https://golang.org/pkg/io/ioutil/#ReadDir
    //
    // https://golang.org/pkg/io/ioutil/#ReadFile
    //

    // Print out a few lines telling the user that the program has started.
    fmt.Println("\n-----------------------------------------------")
    fmt.Println("Hardware Temperature Info Tool for Linux x86-64")
    fmt.Println("-----------------------------------------------\n")

    // Attempt to read in our file contents.
    list_of_device_dirs, err := ioutil.ReadDir(hardware_monitor_directory)
    panic_if_error(err)

    // Debug mode, print out a list of files in the directory specified by
    // the "hardware_monitor_directory" global variable.
    if debug_mode {

        // Tell the end-user we are in debug mode.
        debug_print("The following IDs are present in the hardware sensor " +
                    "monitoring directory:\n")

        // String to hold out concat list of hardware device directories.
        debug_string_for_list_of_device_dirs := ""

        // Cycle thru the array that holds the directories.
        for _, dir := range list_of_device_dirs {
            debug_string_for_list_of_device_dirs += "* " + dir.Name() + "\n"
        }

        // Finally, print out a list of device 
        debug_print(debug_string_for_list_of_device_dirs)
    }

    // For each of the devices...
    for _, dir := range list_of_device_dirs {

        // Assemble the filepath to the name file of the currently given
        // hardware device.
        hardware_name_filepath_of_given_device := hardware_monitor_directory +
          dir.Name() + "/" + hardware_name_file

        // If debug mode, print out the current 'name' file we are about
        // to open.
        debug_print(dir.Name() + " --> " +
          hardware_name_filepath_of_given_device)

        // ...check to see if a 'name' file is present inside the directory.
        name_value_of_hardware_device, err := ioutil.ReadFile(
          hardware_name_filepath_of_given_device)

        // If err is not nil, skip this device.
        if err != nil {

            // If debug mode, then print out a message telling the user
            // which device is missing a hardware 'name' file.
            debug_print("Warning: " + dir.Name() + " does not contain a " +
                        "hardware name file. Skipping...")

            // Move on to the next device.
            continue
        }

        // If the hardware name file does not contain anything of value,
        // skip it and move on to the next device.
        if len(name_value_of_hardware_device) < 1 {

            // If debug mode, then print out a message telling the user
            // which device is missing a hardware 'name' file.
            debug_print("Warning: The hardware name file of " + dir.Name() +
                        " does not contain valid data. Skipping...")

            // Move on to the next device.
            continue
        }

        // Trim away any excess whitespace from the hardware name file data.
        name_value_of_hardware_device_as_string :=
          strings.Trim(string(name_value_of_hardware_device), " \n")

        // Now, since the hardware name file contains valid string data, 
        // go ahead and print out the device name.
        fmt.Println("Device: " + name_value_of_hardware_device_as_string +
                    " (" + dir.Name() + ")")
        fmt.Println("-------")

        // Assemble the filepath to the temperature file of the currently
        // given hardware device.
        hardware_temperature_filepath_of_given_device :=
          hardware_monitor_directory + dir.Name() + "/" + hardware_temp_file

        // If debug mode, print out the current 'temperature' file we are
        // about to open.
        debug_print(dir.Name() + " --> " +
          hardware_temperature_filepath_of_given_device)

        // If the hardware monitor contains an actively-updating temperature
        // sensor, then attempt to open it. 
        temperature_value_of_hardware_device, err := ioutil.ReadFile(
          hardware_temperature_filepath_of_given_device)

        // If err is not nil, or temperature file is empty, tell the 
        // end-user this device does not appear to have temperature data.
        if err != nil || len(temperature_value_of_hardware_device) < 1 {

            // If debug mode, then print out a message telling the user
            // which device is missing a hardware 'temperature' file.
            debug_print("Warning: " + dir.Name() + " does not contain a " +
                        "valid hardware temperature file, ergo no " +
                        "temperature data to print for this device.")

            // Print a none-available since no temperature data is available.
            fmt.Println("\nTemp: N/A\n\n")

            // With that done, go ahead and move on to the next device.
            continue
        }

        // If debug mode, tell the end-user that this is converted to
        // a string.
        debug_print("Converting temperature file data from " +
                    dir.Name() + " into a string.")

        // Attempt to convert the temperature to a string, trim it, and then
        // to an integer value afterwards.
        temperature_value_of_hardware_device_as_int, err :=
          strconv.Atoi(strings.Trim(string(temperature_value_of_hardware_device), " \n"))

        // If err is not nil, then the temperature file does not have valid 
        // integer data. So tell the end-user no data is available.
        if err != nil || len(temperature_value_of_hardware_device) < 1 {

            // If debug mode, then print out a message telling the user
            // which device is missing a hardware 'temperature' file.
            debug_print("Warning: " + dir.Name() + " does not contain a " +
                        "integer data in the hardware temperature file, " +
                        "ergo no temperature data to print for this device.")

            // Print a none-available since no temperature data is available.
            fmt.Println("\nTemp: N/A\n\n")

            // With that done, go ahead and move on to the next device.
            continue
        }

        // Usually hardware sensor is uses 3-sigma of precision and stores
        // the value as an integer for purposes of simplicity.
        //
        // Ergo, this needs to be divided by 1000 to give temperature
        // values that are meaningful to humans.
        //
        temperature_value_of_hardware_device_as_int /= 1000

        // This acts as a work-around for the k10temp sensor module.
        if name_value_of_hardware_device_as_string == "k10temp" {

            // Add 30 degrees to the current temperature.
            temperature_value_of_hardware_device_as_int += 30
        }

        // Finally, print out the temperature data of the current device.
        fmt.Println("\nTemp:", temperature_value_of_hardware_device_as_int, "C\n\n")
    }

    // If all is well, we can return quietly here.
    os.Exit(0)
}
