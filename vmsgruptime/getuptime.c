#include "openvms.h"
#include "dat.h"
#include "fns.h"

uint64_t
getuptime(void)
{
	int64_t ftime;
	int64_t now;
	uint64_t ftimelen;
	int status;
	ItemList64 ftime_desc = {1, SYI$_BOOTTIME, -1, 8, (uint64_t)&ftime, (uint64_t)&ftimelen};

	status = sys$getsyiw(0, 0, 0, &ftime_desc, 0, 0, 0);
	if(!($VMS_STATUS_SUCCESS(status)))
		lib$stop(status);
	status = sys$gettim((void*)&now);
	if(!($VMS_STATUS_SUCCESS(status)))
		lib$stop(status);
	ftime -= now;
	ftime = (ftime*-1)/1000/1000/10;
	return ftime;
}

uint64_t /* copy paste job */
getclusteruptime(void)
{
	int64_t ftime;
	int64_t now;
	int status;
	uint64_t ftimelen;
	ItemList64 ftime_desc = {1, SYI$_CLUSTER_FTIME, -1, 8, (uint64_t)&ftime, (uint64_t)&ftimelen};

	status = sys$getsyiw(0, 0, 0, &ftime_desc, 0, 0, 0);
	if(!($VMS_STATUS_SUCCESS(status)))
		lib$stop(status);
	status = sys$gettim((void*)&now);
	if(!($VMS_STATUS_SUCCESS(status)))
		lib$stop(status);
	ftime -= now;
	ftime = (ftime*-1)/1000/1000/10;
	return ftime;
}

int
isclustermember(void)
{
	char cluster_status;
	int status;
	uint64_t statuslen;
	ItemList64 cs_desc = {1, SYI$_CLUSTER_MEMBER, -1, 1, (uint64_t)&cluster_status, (uint64_t)&statuslen};

	status = sys$getsyiw(0, 0, 0, &cs_desc, 0, 0, 0);
	if(!($VMS_STATUS_SUCCESS(status)))
		lib$stop(status);
	if(cluster_status & 1)
		return 1;
	return 0;
}
