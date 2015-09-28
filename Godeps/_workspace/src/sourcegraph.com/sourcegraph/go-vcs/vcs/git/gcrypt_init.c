#include <gcrypt.h>
#include <errno.h>
#include <pthread.h>
#include <stdio.h>
GCRY_THREAD_OPTION_PTHREAD_IMPL;

int _govcs_gcrypt_init()
{
    int ret;

    // Initialize gcrypt for multithreaded operation. Otherwise
    // concurrent SSH operations crash.
    ret = gcry_control(GCRYCTL_SET_THREAD_CBS, &gcry_threads_pthread);
    if (ret != 0) return ret;
    gcry_check_version((void*)0);

    return 0;
}
