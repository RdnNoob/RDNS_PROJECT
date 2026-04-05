#!/bin/bash
set -e
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

ARCHIVE="rdns_project.tar.gz"
EXPECTED_MD5="ad2885b61eb040a4bcc5c58079b0b43f"
INSTALL_DIR="rdns"
BINARY="rdns_master"

echo -e "${CYAN}  RDNS V2.9 PRO - AUTO INSTALLER${NC}"
echo ""

if [ -d "/data/data/com.termux/files/usr/bin" ]; then
  ENV="termux"
  echo -e "${GREEN}[*] Termux detected${NC}"
elif command -v apt-get >/dev/null 2>&1; then
  ENV="debian"
  echo -e "${GREEN}[*] Debian/Ubuntu detected${NC}"
else
  ENV="other"
  echo -e "${YELLOW}[*] Other environment${NC}"
fi

if command -v go >/dev/null 2>&1; then
  echo -e "${GREEN}[+] golang ok${NC}"
else
  echo -e "${YELLOW}[*] Install golang...${NC}"
  if [ "$ENV" = "termux" ]; then pkg install golang -y
  elif [ "$ENV" = "debian" ]; then apt-get install -y golang
  else echo -e "${RED}[-] Install golang manual${NC}"; exit 1; fi
fi

if command -v ssh >/dev/null 2>&1; then
  echo -e "${GREEN}[+] openssh ok${NC}"
else
  echo -e "${YELLOW}[*] Install openssh...${NC}"
  if [ "$ENV" = "termux" ]; then pkg install openssh -y
  elif [ "$ENV" = "debian" ]; then apt-get install -y openssh-client; fi
fi

if command -v curl >/dev/null 2>&1; then
  echo -e "${GREEN}[+] curl ok${NC}"
else
  echo -e "${YELLOW}[*] Install curl...${NC}"
  if [ "$ENV" = "termux" ]; then pkg install curl -y
  elif [ "$ENV" = "debian" ]; then apt-get install -y curl; fi
fi

echo -e "${YELLOW}[*] Verifikasi MD5...${NC}"
if command -v md5sum >/dev/null 2>&1; then
  ACTUAL=$(md5sum "$ARCHIVE" | awk '{print $1}')
  if [ "$ACTUAL" = "$EXPECTED_MD5" ]; then
    echo -e "${GREEN}[+] MD5 OK: $ACTUAL${NC}"
  else
    echo -e "${RED}[-] MD5 MISMATCH!${NC}"
    echo -e "${RED}    Expected : $EXPECTED_MD5${NC}"
    echo -e "${RED}    Actual   : $ACTUAL${NC}"
    exit 1
  fi
else
  echo -e "${YELLOW}[!] md5sum tidak tersedia, skip${NC}"
fi

echo -e "${YELLOW}[*] Ekstrak $ARCHIVE...${NC}"
if [ ! -f "$ARCHIVE" ]; then
  echo -e "${RED}[-] $ARCHIVE tidak ditemukan!${NC}"
  exit 1
fi
tar -xzf "$ARCHIVE"
if [ ! -d "$INSTALL_DIR" ]; then
  echo -e "${RED}[-] Ekstraksi gagal!${NC}"
  exit 1
fi
echo -e "${GREEN}[+] Ekstrak ke $INSTALL_DIR/${NC}"

cd "$INSTALL_DIR"
mkdir -p database tmp
if [ ! -f go.mod ]; then
  go mod init rdns_suite
fi
echo -e "${YELLOW}[*] go mod tidy...${NC}"
go mod tidy 2>/dev/null || true
echo -e "${YELLOW}[*] Kompilasi...${NC}"
go build -ldflags="-s -w" -o "$BINARY" .
if [ $? -eq 0 ] && [ -f "$BINARY" ]; then
  chmod +x "$BINARY"
else
  echo -e "${RED}[-] Build gagal!${NC}"
  exit 1
fi
cd ..

echo ""
echo -e "${GREEN}[OK] INSTALASI SELESAI!${NC}"
echo -e "${CYAN}  Jalankan : cd $INSTALL_DIR && ./$BINARY${NC}"
echo -e "${CYAN}  Panel    : http://localhost:8080/?admin=rdns${NC}"
