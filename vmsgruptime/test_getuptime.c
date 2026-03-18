#include "openvms.h"
#include "dat.h"
#include "fns.h"

int
main(int argc, char *argv[])
{
	int64_t uptime_seconds;
	int64_t uptime_100ns;
	int64_t uptime_ns;
	unsigned short timelen;
	char timestring[21];
	struct dsc$descriptor timebuf = {
		20,
		DSC$K_DTYPE_T,
		DSC$K_CLASS_S,
		timestring,
	};
	int status;
	int64_t clu_uptime_seconds;
	int64_t clu_uptime_ns;
	int64_t clu_uptime_100ns;
	int cluster_member;
	char clu_timestring[21];
	unsigned short clu_timelen;
	struct dsc$descriptor clu_timebuf = {
		20,
		DSC$K_DTYPE_T,
		DSC$K_CLASS_S,
		clu_timestring,
	};
	struct _generic_64 *up100ns;
	struct _generic_64 *cup100ns;

	uptime_seconds = getuptime();
	uptime_100ns = uptime_seconds*1000*1000*10*-1;
	uptime_ns = uptime_seconds*1000000000;
	if(uptime_seconds == 0)
		panic("unable to get uptime");
	up100ns = (void*)&uptime_100ns;
	status = sys$asctim(&timelen, &timebuf, up100ns, 0);
	if(!(status & 1))
		panic("unable to format uptime");
	timestring[timelen] = 0;
	cluster_member = isclustermember();
	if(cluster_member){
		clu_uptime_seconds = getclusteruptime();
		clu_uptime_100ns = clu_uptime_seconds*1000*1000*10*-1;
		cup100ns = (void*)&clu_uptime_100ns;
		status = sys$asctim(&clu_timelen, &clu_timebuf, cup100ns, 0);
		if(!(status & 1))
			panic("unable to format cluster uptime");
		clu_timestring[clu_timelen] = 0;
	}
	printf("uptime_seconds = %lld\n", uptime_seconds);
	printf("uptime_ns = %lld\n", uptime_ns);
	printf("formatted uptime = \"%s\"\n", timestring);
	printf("cluster_member = %d\n", cluster_member);
	if(cluster_member){
		printf("cluster_uptime_seconds = %lld\n", clu_uptime_seconds);
		printf("formatted cluster uptime = \"%s\"\n", clu_timestring);
	}
	return 0;
}
