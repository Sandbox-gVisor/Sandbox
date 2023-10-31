#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <sys/socket.h>
#include <arpa/inet.h>
#include <errno.h>
#include <unistd.h>

#define FAILURE (-1)

#define ERROR_INVALID_ARGS_COUNT    1
#define ERROR_PORT_CONVERSATION     2
#define ERROR_BIND                  3

// accepts the port to bind
// returns socket fd which is bind to localhost:port or -1 if error occurred
int bind_to_port(int port) {
    int sock_fd = socket(AF_INET, SOCK_STREAM, 0);
    if (sock_fd < 0) {
        perror("socket");
        return FAILURE;
    }
    struct sockaddr_in addr;
    memset(&addr, 0, sizeof(struct sockaddr_in));
    addr.sin_family = AF_INET;
    addr.sin_addr.s_addr = htonl(INADDR_ANY);
    addr.sin_port = htons(port);

    int bind_res = bind(sock_fd, (struct sockaddr*)&addr, sizeof(struct sockaddr_in));
    if (bind_res != 0) {
        perror("bind");
        int close_res = close(sock_fd);
        if (close_res < 0) {
            perror("close sock_fd");
        }
        return FAILURE;
    }

    return sock_fd;
}

int main(int argc, char* argv[]) {
    if (argc < 2) {
        fprintf(stderr, "Error wrong amount of arguments\n");
        exit(ERROR_INVALID_ARGS_COUNT);
    }
    char *invalid_sym;
    errno = 0;
    int port = (int)strtol(argv[1], &invalid_sym, 10);
    if (errno != 0 || *invalid_sym != '\0') {
        fprintf(stderr, "Error wrong port\n");
        exit(ERROR_PORT_CONVERSATION);
    }

    int listen_fd = bind_to_port(port);
    if (listen_fd == FAILURE) {
        exit(ERROR_BIND);
    }
    printf("\nSuccessfully bind to port %d :)\n\n", port);
}
