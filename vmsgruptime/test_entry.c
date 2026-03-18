#include <starlet.h>
#include <syidef.h>
#include <descrip.h>
#include <stdio.h>
#include <stdlib.h>
#include <types.h>
#include <string.h>
#include <in.h>
#include "dat.h"
#include "fns.h"

extern int test_encode(int argc, char *argv[]);
extern int test_getuptime(int argc, char *argv[]);
extern long long int WHICH_TEST;

int
main(int argc, char *argv[])
{
	switch(WHICH_TEST){
	case 1:
		return test_encode(argc, argv);
	case 2:
		return test_getuptime(argc, argv);
	default:
		fprintf(stderr, "test %llu is not defined", WHICH_TEST);
		return -1;
	}
	return -1;
}
