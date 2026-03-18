#include "openvms.h"
#include "dat.h"
#include "fns.h"

void
panic(char *s)
{
	fprintf(stderr, "panic: %s\n", s);
	exit(-1);
}

void
warning(char *s)
{
	fprintf(stderr, "warning: %s\n", s);
}

void*
emallocz(size_t sz)
{
	void *ptr;

	ptr = malloc(sz);
	if(!ptr)
		panic("bad malloc");
	memset(ptr, 0, sz);
	return ptr;
}

uint64_t
swapbytes(uint64_t n)
{
	unsigned char *c = (unsigned char*)&n;
	uint64_t res;

	res = ((uint64_t)c[0] << 56) | ((uint64_t)c[1] << 48) | ((uint64_t)c[2] << 40) |
			((uint64_t)c[3] << 32) | ((uint64_t)c[4] << 24) | ((uint64_t)c[5] << 16) |
			((uint64_t)c[6] << 8) | (uint64_t)c[7];
	return res;
}

uint64_t
htonll(uint64_t n)
{
	return swapbytes(n);
}

uint64_t
ntohll(uint64_t n)
{
	return swapbytes(n);
}
