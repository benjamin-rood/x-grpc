package main

import "fmt"

/*
*

	note: since the saved file has no guaranteed file size limit* (hypothetically it could
	be greater than the availability of the available memory), the only way to prevent
	crashing by running out of memory is to either assert an upper file size limit
	beneath the currently availble memory on the system, or, we must do an on-disk
	byte traversal of the JSON tree. This is not part of the assignment, so we will just
	error out if the JSON file to be loaded exceeds available memory.
*/
func modifyJSON() error {
	return fmt.Errorf("not implemented")
}
