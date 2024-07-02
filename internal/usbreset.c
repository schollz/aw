/* usbreset -- send a USB port reset to a USB device */

#include <errno.h>
#include <fcntl.h>
#include <linux/usbdevice_fs.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/ioctl.h>
#include <unistd.h>

#define BUFFER_SIZE 256

int find_stm_device(char *device_filename) {
  FILE *fp;
  char path[BUFFER_SIZE];
  char bus[4], device[4];
  char *command = "lsusb | grep 'STMicroelectronics'";
  char *prefix = "/dev/bus/usb/";

  fp = popen(command, "r");
  if (fp == NULL) {
    perror("Failed to run lsusb command");
    return 1;
  }

  if (fgets(path, sizeof(path) - 1, fp) != NULL) {
    sscanf(path, "Bus %3s Device %3s", bus, device);
    snprintf(device_filename, BUFFER_SIZE, "%s%s/%s", prefix, bus, device);
  } else {
    fprintf(stderr, "No STMicroelectronics device found\n");
    pclose(fp);
    return 1;
  }

  pclose(fp);
  return 0;
}

int main() {
  char device_filename[BUFFER_SIZE];
  int fd;
  int rc;

  if (find_stm_device(device_filename)) {
    return 1;
  }

  fd = open(device_filename, O_WRONLY);
  if (fd < 0) {
    perror("Error opening output file");
    return 1;
  }

  printf("Resetting USB device %s\n", device_filename);
  rc = ioctl(fd, USBDEVFS_RESET, 0);
  if (rc < 0) {
    perror("Error in ioctl");
    return 1;
  }
  printf("Reset successful\n");

  close(fd);
  return 0;
}
