#include "openvms.h"
#include "dat.h"
#include "fns.h"

int
main(int argc, char *argv[])
{
	int32_t ijoblim = 0;
	uint64_t nusers = 0;
	Userlist *users;
	int i;

	ijoblim = getijoblim();
	nusers = getnusers();
	printf("ijoblim = %d\n", ijoblim);
	printf("nusers = %llu\n", nusers);
	users = getuserlist();
	for(i = 0; i < users->nuser; i++)
		printf("user = \"%s\", nprocs = %d\n",
			users->users[i].username, users->users[i].nprocs);
	freeuserlist(users);
	return 0;
}
