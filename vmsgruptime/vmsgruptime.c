#include "openvms.h"
#include "dat.h"
#include "fns.h"

#define MulticastAddr "239.77.86.0"
#define MulticastPort 3825

int
main(int argc, char *argv[])
{
	int sendsock;
	struct sockaddr_in mcastaddr;
	Uptime *uptime;
	char *msg;
	int msglen;
	int nbytes;

	printf("OpenVMS gruptime status broadcaster\n");
	if((sendsock = socket(AF_INET, SOCK_DGRAM, 0)) < 0){
		perror("socket");
		return -1;
	}

	memset(&mcastaddr, 0, sizeof(mcastaddr));
	mcastaddr.sin_family = AF_INET;
	mcastaddr.sin_addr.s_addr = inet_addr(MulticastAddr);
	mcastaddr.sin_port = htons(MulticastPort);

	for(;;){
		uptime = getdata();
		msg = encodemessage(uptime);
		msglen = msg[0];
		nbytes = sendto(sendsock, msg, msglen, 0, (struct sockaddr*)&mcastaddr, sizeof(mcastaddr));
		if(nbytes < 0){
			perror("sendto");
			return -2;
		}
		free(uptime->hostname);
		free(uptime);
		free(msg);
		sleep(LIFETIME/2);
	}

	return 0;
}

