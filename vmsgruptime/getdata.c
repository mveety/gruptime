#include "openvms.h"
#include "dat.h"
#include "fns.h"

Uptime*
getdata(void)
{
	Uptime *uptime;
	char sysnamebuf[18];

	uptime = emallocz(sizeof(Uptime));
	getsysname((char*)&sysnamebuf[0], sizeof(sysnamebuf));
	uptime->hostname = strdup((char*)&sysnamebuf[0]);
	uptime->ostype = OPENVMS;
	uptime->uptime = getuptime()*1000000000;
	uptime->load1 = 0;
	uptime->load5 = 0;
	uptime->load15 = 0;
	uptime->nusers = getnusers();
	uptime->lifetime = ((int64_t)LIFETIME)*1000000000;
	return uptime;
}

void
freeuptime(Uptime *uptime)
{
	free(uptime->hostname);
	free(uptime);
}

