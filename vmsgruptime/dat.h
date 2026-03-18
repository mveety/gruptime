/* OpenVMS gruptime server */

enum {
	/* operating systems */
	UNKNOWN = 254,
	FREEBSD = 1,
	LINUX = 2,
	WINDOWS = 3,
	OPENVMS = 4,
	PLAN9 = 9,
	/* other stuff */
	PROTOVERSION = 4,
	LIFETIME = 480,
	UDPPORT = 3825,
	TCPPORT = 3826,
	MAXPEERS = 256,
};

typedef struct Uptime Uptime;
typedef struct ItemList64 ItemList64;
typedef struct ItemList ItemList;
typedef struct User User;
typedef struct Userlist Userlist;
typedef struct Peer Peer;
typedef struct Peers Peers;
typedef struct String String;

struct Uptime {
	char *hostname;
	char ostype;
	uint64_t uptime;
	double load1;
	double load5;	
	double load15;
	uint64_t nusers;
	uint64_t lifetime;
};

struct ItemList64 {
	uint16_t MBO;
	uint16_t code;
	int32_t MBMO;
	uint64_t len;
	uint64_t bufaddr;
	uint64_t retaddr;
};

struct ItemList {
	uint16_t len;
	uint16_t code;
	uint32_t bufaddr;
	uint32_t flags;
};

struct User {
	char username[13];
	uint32_t nprocs;
};

struct Userlist {
	int32_t maxusers;
	int32_t nuser;
	User *users;
};

struct Peer {
	char *hostname;
};

struct Peers {
	uint32_t npeers;
	uint32_t peerssz;
	Peer **peers;
};

struct String {
	size_t len;
	char *data;
};

