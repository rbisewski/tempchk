# tempchk - Hardware temperature sensor tool for Linux, written in golang.

This is a rather plain golang-based method of gathering temperature data from
common hardware sensors on Linux.

When it comes to temperature data on Linux, I expect there is a great deal
of non-conformity, so go ahead and shoot me an email if you see any obvious
mistakes on your hardware.

Part of my reason for writting this was a need for a more minimalist version
of the current solution, lm-sensors, which perhaps a bit of a heavy-weight
implementation.

Maybe one day it will be more fleshed out, but for now it is more of a
simple tool. Feel free to fork it and use it for other projects if you find
it useful.


# Requirements

The following is needed in order for this to function as intended:

* Linux kernel 4.4+
* golang

Older kernels could still give some kind of result, but I *think* most of
the newer Linux distros have since made the /proc method of retrieving the
temperature obsolete. Feel free to correct me if this is incorrect.


# Running

0) Build this program as you would a simple golang module.

go build tempchk

1) Run the program.

./tempchk


# Authors

Written by Robert Bisewski at Ibis Cybernetics. For more information, contact:

* Website -> www.ibiscybernetics.com

* Email -> contact@ibiscybernetics.com
