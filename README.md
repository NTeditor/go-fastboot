## Example

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nteditor/go-fastboot"
	"github.com/nteditor/go-fastboot/fastbootErrors"
)

func main() {
	host, closeHost := fastboot.NewHost()
	defer closeHost()
	devs, err := host.ListDevices()
	if err != nil {
		fmt.Println("Error listing devices:", err)
		return
	}
	if len(devs) == 0 {
		fmt.Println("No devices found.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

// Reboot device
	err = devs[0].Reboot(ctx)
	if err != nil {
		if errors.Is(err, fastbootErrors.ErrTimeout) {
			fmt.Println("Reboot timed out.")
		} else if errFail, ok := err.(*fastbootErrors.ErrStatusFail); ok {
			fmt.Println("Reboot failed with status:", errFail.Data)
		} else if errors.Is(err, fastbootErrors.ErrDeviceClose) {
			fmt.Println("Device descriptor closed.")
		} else {
			fmt.Println("Reboot error:", err)
		}
	}

}
```

## Problems

- Bootloader mode may not work correctly in Windows

## Licensing

This project is licensed under the GNU General Public License v3.0.

It also includes components licensed under the Apache License 2.0, specifically the `gousb` library by Google.

## Acknowledgements

This project uses the following third-party libraries:

- [gousb](https://github.com/google/gousb) by Google, licensed under the Apache License 2.0.

