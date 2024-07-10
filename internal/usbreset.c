
#include <errno.h>
#include <fcntl.h>
#include <linux/usbdevice_fs.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/ioctl.h>
#include <unistd.h>

#define BUFFER_SIZE 256

int find_stm_devices(char device_filenames[][BUFFER_SIZE], int max_devices) {
  FILE *fp;
  char path[BUFFER_SIZE];
  char bus[4], device[4];
  char *command = "lsusb | grep 'STMicroelectronics'";
  char *prefix = "/dev/bus/usb/";
  int count = 0;

  fp = popen(command, "r");
  if (fp == NULL) {
    perror("Failed to run lsusb command");
    return 1;
  }

  while (fgets(path, sizeof(path) - 1, fp) != NULL && count < max_devices) {
    sscanf(path, "Bus %3s Device %3s", bus, device);
    snprintf(device_filenames[count], BUFFER_SIZE, "%s%s/%s", prefix, bus,
             device);
    count++;
  }

  pclose(fp);
  return count;
}

int main() {
  char device_filenames[10][BUFFER_SIZE];  // Adjust the array size as needed
  int num_devices;
  int fd;
  int rc;

  num_devices = find_stm_devices(device_filenames, 10);
  if (num_devices == 0) {
    fprintf(stderr, "No STMicroelectronics device found\n");
    return 1;
  }

  for (int i = 0; i < num_devices; i++) {
    fd = open(device_filenames[i], O_WRONLY);
    if (fd < 0) {
      perror("Error opening output file");
      continue;
    }

    printf("Resetting USB device %s\n", device_filenames[i]);
    rc = ioctl(fd, USBDEVFS_RESET, 0);
    if (rc < 0) {
      perror("Error in ioctl");
      close(fd);
      continue;
    }
    printf("Reset successful\n");

    close(fd);
  }

  return 0;
}