PROG      := dhcpscript-sender
OPTS      := -DDO_COUNT -DTEXT_REPLY -DREAD_FILE -DREAD_GIF -DNULLSERV_REPLIES -DSSL_RESP
TEST_OPTS := -DTEST -DVERBOSE
TINY_OPTS := -O3 -DTINY
DEBUG_OPT := -DHEX_DUMP
CC        := gcc
CFLAGS    += -Os -s -Wall -ffunction-sections -fdata-sections -fno-strict-aliasing
LDFLAGS   += -Wl,--gc-sections
STRIP     := strip -s -R .note -R .comment -R .gnu.version -R .gnu.version_r
CROSSCC   := mipsel-uclibc-gcc
CROSSSTRIP:= mipsel-uclibc-$(STRIP)
SRC        = $(PROG).c

# mips uclib toolchain
export PATH := /opt/brcm/hndtools-mipsel-uclibc/bin:/opt/brcm/hndtools-mipsel-linux/bin:$(PATH)

all: dist mips tiny host32 host64
	@echo "Builds in dist folder."

dist:
	@mkdir dist

mips:
	$(CROSSCC) -mips32 $(CFLAGS) $(LDFLAGS) $(OPTS) $(SRC) -o dist/$(PROG).mips32
	$(CROSSSTRIP) dist/$(PROG).mips32
	upx dist/$(PROG).mips32

host32:
	$(CC) -m32 $(CFLAGS) $(LDFLAGS) $(OPTS) $(SRC) -o dist/$(PROG).x86
	strip dist/$(PROG).x86

host64:
	$(CC) -m64 $(CFLAGS) $(LDFLAGS) $(OPTS) $(SRC) -o dist/$(PROG).amd64
	strip dist/$(PROG).amd64

tiny:
	$(CROSSCC) -mips32 $(CFLAGS) $(LDFLAGS) $(TINY_OPTS) $(SRC) -o dist/$(PROG).tiny.mips32
	$(CROSSSTRIP) dist/$(PROG).tiny.mips32
	upx dist/$(PROG).tiny.mips32
