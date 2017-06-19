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

    // Flag to check if the device in question uses the 'fam15h_power'
    // kernel module, in which case the temperature adjustment phase can be
    // skipped if the 'k10temp' module is also present.
    fam15h_power_module_in_use = false

    // size of the longest hwmonX/name entry string
    max_entry_length = 0

    // spacer size
    spacer_size = 4
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

    // Search thru the directories and set the relevant flags...
    err = SetGlobalSensorFlags(list_of_device_dirs)

    // safety check, ensure no errors occurred
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
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

            // append string values equivalent to the longest length.
            for len(name_value_of_hardware_device_as_string) <
              max_entry_length+spacer_size {

                // append a space to the end of the string
                name_value_of_hardware_device_as_string += " "
            }

            // Print a none-available since no temperature data is available.
            fmt.Println(dir.Name(), " | ",
              name_value_of_hardware_device_as_string, "N/A\n")

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

            // append string values equivalent to the longest length.
            for len(name_value_of_hardware_device_as_string) <
              max_entry_length+spacer_size {

                // append a space to the end of the string
                name_value_of_hardware_device_as_string += " "
            }

            // Finally, print out the temperature data of the current device.
            fmt.Println(dir.Name(), " | ",
              name_value_of_hardware_device_as_string, "N/A\n")

            // With that done, go ahead and move on to the next device.
            continue
        }

        // Usually hardware sensors uses 3-sigma of precision and stores
        // the value as an integer for purposes of simplicity.
        //
        // Ergo, this needs to be divided by 1000 to give temperature
        // values that are meaningful to humans.
        //
        temperature_value_of_hardware_device_as_int /= 1000

        // This acts as a work-around for the k10temp sensor module.
        if name_value_of_hardware_device_as_string == "k10temp" &&
          !fam15h_power_module_in_use {

            // Add 30 degrees to the current temperature.
            temperature_value_of_hardware_device_as_int += 30
        }

        // append string values equivalent to the longest length.
        for len(name_value_of_hardware_device_as_string) <
          max_entry_length+spacer_size {

            // append a space to the end of the string
            name_value_of_hardware_device_as_string += " "
        }

        // Finally, print out the temperature data of the current device.
        fmt.Println(dir.Name(), " | ",
          name_value_of_hardware_device_as_string,
          temperature_value_of_hardware_device_as_int, "C\n")
    }

    // If all is well, we can return quietly here.
    os.Exit(0)
}

//! Set global flags which may alter how Linux seems temperatures
/*
 * @param    os.FileInfo[]    array of directory data
 *
 * @return   error            error message, if any
 */
func SetGlobalSensorFlags(dirs []os.FileInfo) (error) {

    // input validation
    if dirs == nil || len(dirs) < 1 {
        return fmt.Errorf("SetGlobalSensorFlags() --> invalid input")
    }

    // Cycle thru the entire list of device directories...
    for _, dir := range dirs {

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

        // Determine the length of the longest entry string
        //
        // TODO: this is a less than ideal place for this code, consider
        //       rewriting how this program handles hwmonX entries at some
        //       future date
        //
        if len(name_value_of_hardware_device_as_string) > max_entry_length {
            max_entry_length = len(name_value_of_hardware_device_as_string)
        }

        // Conduct a quick check to determine if the 'fam15h_power' module
        // is currently in use.
        if name_value_of_hardware_device_as_string == "fam15h_power" {
            fam15h_power_module_in_use = true
        }
    }

    // everything worked fine, so return null
    return nil
}
