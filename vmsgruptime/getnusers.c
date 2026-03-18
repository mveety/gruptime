#include "openvms.h"
#include "dat.h"
#include "fns.h"

/* max users is SYI$_IJOBLIM */

int32_t
getijoblim(void)
{
	int32_t ijoblim;
	ItemList64 ijoblim_desc = {1, SYI$_IJOBLIM, -1, 4, (uint64_t)&ijoblim, 0};
	uint32_t status;

	status = sys$getsyiw(0, 0, 0, &ijoblim_desc, 0, 0, 0);
	if(!($VMS_STATUS_SUCCESS(status)))
		lib$stop(status);
	return ijoblim;
}

Userlist*
getuserlist(void)
{
	uint32_t status;
	uint64_t procs_count = 0;
	uint32_t mode = 1;
	uint32_t pid = 0;
	IOSB iostat;
	uint32_t psctx = 0;
	char username[13];
	const char *login_user = "<login>     ";
	Userlist *users;
	int i;
	ItemList pscan_desc[] = {
		{0, PSCAN$_MODE, JPI$K_INTERACTIVE, PSCAN$M_EQL},
		{0, 0, 0, 0},
	};
	ItemList getjpi_desc[] = {
		{12, JPI$_USERNAME, (uint32_t)&username, 0},
		{4, JPI$_MODE, (uint32_t)&mode, 0},
		{0, 0, 0, 0},
	};

	users = emallocz(sizeof(Userlist));
	users->maxusers = getijoblim();
	users->nuser = 0;
	users->users = emallocz(sizeof(User)*users->maxusers);
	status = sys$process_scan(&psctx, &pscan_desc);
	if(!($VMS_STATUS_SUCCESS(status)))
		lib$stop(status);
	for(;;) {
		memset(&username, 0, sizeof(username));
		status = sys$getjpiw(EFN$C_ENF, &psctx, 0, &getjpi_desc, &iostat, 0, 0);
		if(iostat.iosb$l_getxxi_status == SS$_NOMOREPROC)
			return users;
		if(!($VMS_STATUS_SUCCESS(status)))
			lib$stop(status);
		username[12] = 0;
		if(strcmp((char*)&username, login_user) == 0)
			continue;
		for(i = 0; i < sizeof(username); i++)
			if(username[i] == ' ')
				username[i] = 0;
		if(users->nuser == 0) {
			memcpy(&users->users[0].username, &username, sizeof(username));
			users->users[0].nprocs = 1;
			users->nuser++;
		} else {
			for(i = 0; i < users->nuser; i++){
				if(strcmp((char*)&users->users[i].username, (char*)&username) == 0){
					users->users[i].nprocs++;
					break;
				}
			}
			if(i == users->nuser && i < users->maxusers){
				memcpy(&users->users[i].username, &username, sizeof(username));
				users->users[i].nprocs = 1;
				users->nuser++;
			}
		}
		procs_count++;
	}
	free(users->users);
	free(users);
	return 0;
}

void
freeuserlist(Userlist *ul)
{
	free(ul->users);
	free(ul);
}

uint64_t
getnusers(void)
{
	Userlist *userlist;
	uint64_t nusers;

	userlist = getuserlist();
	nusers = userlist->nuser;
	freeuserlist(userlist);
	return nusers;
}
