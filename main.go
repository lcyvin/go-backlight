package main

import (
	"fmt"
	"os"
  "io"
  "strconv"
  "strings"
  "math"

	"github.com/fsnotify/fsnotify"
)

var backlight string
var maxBrightness float64

func TryDefaultBacklight() (string, error) {
  default_backlight_paths := []string{
    "/sys/class/backlight/amdgpu_bl1",
    "/sys/class/backlight/intel_backlight",
    "/sys/class/backlight/acpi_video0",
  }

  found := ""

  for _,v := range default_backlight_paths {
    _,err := os.Stat(v)
    if err == nil {
      found = v
      break
    }
  }

  if found == "" {
    return "", fmt.Errorf("No backlight found from default set, try passing the specific backlight device path")
  }

  return found, nil
}

func PrintBrightnessPercentage(val float64) {
  fmt.Printf("%d%%\n", int(math.Ceil((val/maxBrightness)*100)))
}

func WatchBacklight() error {
  watcher, err := fsnotify.NewWatcher()
  if err != nil {
    return err
  }

  watcher.Add(backlight)
  
  for {
    select {
    case event := <- watcher.Events:
      if event.Name == backlight+"/brightness" {
        val, err := ReadBrightness(backlight+"/brightness")
        if err != nil {
          return err
        }
        PrintBrightnessPercentage(val)
      }
    case err := <- watcher.Errors:
      fmt.Println(err)
      os.Exit(1)
    }
  }
}

func ReadBrightness(fp string) (float64, error) {
  f, err := os.Open(fp)
  if err != nil {
    return 0, fmt.Errorf("Error reading brightness: %v", err)
  }

  v, err := io.ReadAll(f)
  if err != nil {
    return 0, fmt.Errorf("Error reading brightness: %v", err)
  }

  b, err := strconv.ParseFloat(strings.TrimSpace(string(v)), 64)
  if err != nil {
    return 0, fmt.Errorf("Error reading brightness: %v", err)
  }

  return b, nil
}

func init() {
  if b := os.Getenv("BACKLIGHT_DEVICE_PATH"); b == "" {
    b, err := TryDefaultBacklight()
    if err != nil {
      fmt.Println(err)
      os.Exit(1)
    }

    backlight = b
  }

  mb, err := ReadBrightness(fmt.Sprintf("%s/%s", backlight, "max_brightness"))
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  maxBrightness = mb
}


func main() {
  initialBrightness,err := ReadBrightness(fmt.Sprintf("%s/%s", backlight, "brightness"))
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  PrintBrightnessPercentage(initialBrightness)
  WatchBacklight()
}
