TARGETS = mountkernfs.sh hostname.sh mountdevsubfs.sh procps hwclock.sh checkroot.sh urandom bootmisc.sh mountall-bootclean.sh mountall.sh checkroot-bootclean.sh mountnfs-bootclean.sh mountnfs.sh checkfs.sh
INTERACTIVE = checkroot.sh checkfs.sh
mountdevsubfs.sh: mountkernfs.sh
procps: mountkernfs.sh
hwclock.sh: mountdevsubfs.sh
checkroot.sh: hwclock.sh mountdevsubfs.sh hostname.sh
urandom: hwclock.sh
bootmisc.sh: mountall-bootclean.sh checkroot-bootclean.sh mountnfs-bootclean.sh
mountall-bootclean.sh: mountall.sh
mountall.sh: checkfs.sh checkroot-bootclean.sh
checkroot-bootclean.sh: checkroot.sh
mountnfs-bootclean.sh: mountnfs.sh
checkfs.sh: checkroot.sh
