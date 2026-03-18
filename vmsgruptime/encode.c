#include "openvms.h"
#include "dat.h"
#include "fns.h"

char*
byte2os(char ostype)
{
	switch(ostype){
	case FREEBSD: return "FreeBSD";
	case LINUX: return "Linux";
	case WINDOWS: return "Windows";
	case OPENVMS: return "OpenVMS";
	case PLAN9: return "Plan 9";
	case UNKNOWN: return "Unknown";
	default: return "Invalid";
	}
}

char*
encodemessage(Uptime *uptime)
{
	size_t hostlen;
	size_t msglen;
	size_t i = 0;
	uint64_t tmp;
	char *msgbytes;

	char *msglen_bytes;
	char *osbyte;
	char *protoversion;
	char *timebytes;
	char *hostbytes;
	char *load1bytes;
	char *load5bytes;
	char *load15bytes;
	char *nusersbytes;
	char *lifetimebytes;

	hostlen = strlen(uptime->hostname);
	msglen = hostlen+1+1+1+8+8+8+8+8+8;
	msgbytes = emallocz(msglen);

	msglen_bytes = &msgbytes[i++];
	osbyte = &msgbytes[i++];
	protoversion = &msgbytes[i++];
	timebytes = &msgbytes[i]; i += 8;
	hostbytes = &msgbytes[i]; i += hostlen;
	load1bytes = &msgbytes[i]; i += 8;
	load5bytes = &msgbytes[i]; i += 8;
	load15bytes = &msgbytes[i]; i += 8;
	nusersbytes = &msgbytes[i]; i += 8;
	lifetimebytes = &msgbytes[i];

	*msglen_bytes = (char)msglen;
	*osbyte = uptime->ostype;
	*protoversion = PROTOVERSION;
	tmp = htonll(uptime->uptime);
	memcpy(timebytes, &tmp, sizeof(uint64_t));
	memcpy(hostbytes, uptime->hostname, hostlen);
	memcpy(&tmp, &uptime->load1, sizeof(double));
	tmp = htonll(tmp);
	memcpy(load1bytes, &tmp, sizeof(uint64_t));
	memcpy(&tmp, &uptime->load5, sizeof(double));
	tmp = htonll(tmp);
	memcpy(load5bytes, &tmp, sizeof(uint64_t));
	memcpy(&tmp, &uptime->load15, sizeof(double));
	tmp = htonll(tmp);
	memcpy(load15bytes, &tmp, sizeof(uint64_t));
	tmp = htonll(uptime->nusers);
	memcpy(nusersbytes, &tmp, sizeof(uint64_t));
	tmp = htonll(uptime->lifetime);
	memcpy(lifetimebytes, &tmp, sizeof(uint64_t));

	return msgbytes;
}

Uptime*
decodemessage(char *msgbytes)
{
	Uptime *uptime;
	size_t hostlen;
	size_t msglen;
	size_t i = 0;
	uint64_t tmp;

	char *msglen_bytes;
	char *osbyte;
	char *protoversion;
	char *timebytes;
	char *hostbytes;
	char *load1bytes;
	char *load5bytes;
	char *load15bytes;
	char *nusersbytes;
	char *lifetimebytes;

	uptime = emallocz(sizeof(Uptime));

	if(msgbytes[2] < PROTOVERSION){
		warning("protocol too old!");
		return NULL;
	}

	msglen = msgbytes[0];
	hostlen = msglen-(1+1+1+8+8+8+8+8+8);

	msglen_bytes = &msgbytes[i++];
	osbyte = &msgbytes[i++];
	protoversion = &msgbytes[i++];
	timebytes = &msgbytes[i]; i += 8;
	hostbytes = &msgbytes[i]; i += hostlen;
	load1bytes = &msgbytes[i]; i += 8;
	load5bytes = &msgbytes[i]; i += 8;
	load15bytes = &msgbytes[i]; i += 8;
	nusersbytes = &msgbytes[i]; i += 8;
	lifetimebytes = &msgbytes[i];
	
	uptime->hostname = emallocz(hostlen+1);
	memcpy(uptime->hostname, hostbytes, hostlen);	
	uptime->ostype = *osbyte;
	memcpy(&tmp, timebytes, sizeof(uint64_t));
	uptime->uptime = ntohll(tmp);
	memcpy(&tmp, load1bytes, sizeof(uint64_t));
	tmp = ntohll(tmp);
	memcpy(&uptime->load1, &tmp, sizeof(double));
	memcpy(&tmp, load5bytes, sizeof(uint64_t));
	tmp = ntohll(tmp);
	memcpy(&uptime->load5, &tmp, sizeof(double));
	memcpy(&tmp, load15bytes, sizeof(uint64_t));
	tmp = ntohll(tmp);
	memcpy(&uptime->load15, &tmp, sizeof(double));
	memcpy(&tmp, nusersbytes, sizeof(uint64_t));
	uptime->nusers = ntohll(tmp);
	memcpy(&tmp, nusersbytes, sizeof(uint64_t));
	uptime->lifetime = ntohll(tmp);

	return uptime;
}
