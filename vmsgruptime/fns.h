/* util.c */
extern void panic(char*);
extern void warning(char*);
extern void* emallocz(size_t);
extern uint64_t htonll(uint64_t);
extern uint64_t ntohll(uint64_t);

/* encode.c */
extern char* byte2os(char);
extern char* encodemessage(Uptime*);
extern Uptime* decodemessage(char*);

/* getuptime.c */
extern uint64_t getuptime(void);
extern uint64_t getclusteruptime(void);
extern int isclustermember(void);

/* getsysname.c */
extern size_t getsysname(char*, size_t);

/* getnusers.c */
extern int32_t getijoblim(void);
extern Userlist* getuserlist(void);
extern void freeuserlist(Userlist*);
extern uint64_t getnusers(void);

/* getdata.c */
extern Uptime* getdata(void);
extern void freeuptime(Uptime*);


