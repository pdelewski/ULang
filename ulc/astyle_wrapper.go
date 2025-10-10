package main

/*
#cgo CXXFLAGS: -I./astyle -DASTYLE_LIB -std=c++17
#cgo LDFLAGS: -L./astyle -lastyle -lstdc++

#ifdef __cplusplus
extern "C" {
#endif

#include <stdlib.h>

// Forward declarations from astyle
typedef void (*fpError)(int errorNumber, const char* errorMessage);
typedef char* (*fpAlloc)(unsigned long memoryNeeded);

char* AStyleMain(const char* pSourceIn, const char* pOptions, fpError fpErrorHandler, fpAlloc fpMemoryAlloc);
const char* AStyleGetVersion(void);

// Error handler callback
void errorHandler(int errorNumber, const char* errorMessage) {
    // For now, we'll ignore errors or could log them
}

// Memory allocation callback
char* memoryAlloc(unsigned long memoryNeeded) {
    return (char*)malloc(memoryNeeded);
}

#ifdef __cplusplus
}
#endif
*/
import "C"

import (
	"fmt"
	"io/ioutil"
	"log"
	"unsafe"
)

// FormatCodeWithAStyle formats code using the astyle C library
func FormatCodeWithAStyle(sourceCode, options string) (string, error) {
	// Convert Go strings to C strings
	cSourceCode := C.CString(sourceCode)
	defer C.free(unsafe.Pointer(cSourceCode))

	cOptions := C.CString(options)
	defer C.free(unsafe.Pointer(cOptions))

	// Call AStyleMain with our callbacks
	cResult := C.AStyleMain(
		cSourceCode,
		cOptions,
		C.fpError(C.errorHandler),
		C.fpAlloc(C.memoryAlloc),
	)

	if cResult == nil {
		return "", fmt.Errorf("astyle formatting failed")
	}

	// Convert result back to Go string
	result := C.GoString(cResult)

	// Free the memory allocated by astyle
	C.free(unsafe.Pointer(cResult))

	return result, nil
}

// FormatFile formats a single file using astyle
func FormatFile(filePath, options string) error {
	// Read the file
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	// Format the content
	formattedContent, err := FormatCodeWithAStyle(string(content), options)
	if err != nil {
		return fmt.Errorf("failed to format file %s: %v", filePath, err)
	}

	// Write the formatted content back to the file
	err = ioutil.WriteFile(filePath, []byte(formattedContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write formatted file %s: %v", filePath, err)
	}

	log.Printf("Successfully formatted: %s\n", filePath)
	return nil
}

// GetAStyleVersion returns the version of astyle library
func GetAStyleVersion() string {
	cVersion := C.AStyleGetVersion()
	return C.GoString(cVersion)
}
