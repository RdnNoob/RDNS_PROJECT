#!/bin/bash
set -e
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
CYAN='\033[0;36m'; BLUE='\033[0;34m'; NC='\033[0m'
ARCHIVE="rdns_project.tar.gz"
EXPECTED_MD5="ad2885b61eb040a4bcc5c58079b0b43f"
INSTALL_DIR="rdns"
BINARY="rdns_master"

detect_env(){
  if [ -d "/data/data/com.termux/files/usr/bin" ]; then
    ENV="termux"
  elif command -v apt-get &>/dev/null; then
    ENV="debian"
  elif command -v yum &>/dev/null; then
    ENV="redhat"
  else
    ENV="other"
  fi
}

install_pkg(){
  local P=$1 C=${2:-$1}
  command -v "$C" &>/dev/null && { echo -e "${GREEN}[+] $P ok${NC}"; return; }
  echo -e "${YELLOW}[*] Install $P...${NC}"
  case "$ENV" in
    termux) pkg install "$P" -y ;;
    debian) apt-get install -y "$P" ;;
    redhat) yum install -y "$P" ;;
    *) echo -e "${RED}[-] Install $P manual${NC}"; exit 1 ;;
  esac
}

verify_md5(){
  echo -e "${YELLOW}[*] Verifikasi MD5...${NC}"
  if command -v md5sum &>/dev/null; then
    A=$(md5sum "$ARCHIVE" | awk '{print $1}')
  elif command -v md5 &>/dev/null; then
    A=$(md5 -q "$ARCHIVE")
  else
    echo -e "${YELLOW}[!] Skip verifikasi${NC}"; return
  fi
  if [ "$A" = "$EXPECTED_MD5" ]; then
    echo -e "${GREEN}[+] MD5 OK: $A${NC}"
  else
    echo -e "${RED}[-] MD5 MISMATCH!${NC}"
    echo -e "${RED}    Expected : $EXPECTED_MD5${NC}"
    echo -e "${RED}    Actual   : $A${NC}"
    exit 1
  fi
}

extract_archive(){
  echo -e "${YELLOW}[*] Ekstrak $ARCHIVE...${NC}"
  [ -f "$ARCHIVE" ] || { echo -e "${RED}[-] $ARCHIVE tidak ditemukan!${NC}"; exit 1; }
  tar -xzf "$ARCHIVE"
  [ -d "$INSTALL_DIR" ] || { echo -e "${RED}[-] Ekstraksi gagal!${NC}"; exit 1; }
  echo -e "${GREEN}[+] Ekstrak ke $INSTALL_DIR/${NC}"
}

build_binary(){
  cd "$INSTALL_DIR"
  mkdir -p database tmp
  [ -f go.mod ] || go mod init rdns_suite
  echo -e "${YELLOW}[*] go mod tidy...${NC}"
  go mod tidy 2>/dev/null || true
  echo -e "${YELLOW}[*] Kompilasi...${NC}"
  go build -ldflags="-s -w" -o "$BINARY" .
  if [ $? -eq 0 ] && [ -f "$BINARY" ]; then
    chmod +x "$BINARY"
    echo -e "${GREEN}[+] Build berhasil!${NC}"
  else
    echo -e "${RED}[-] Build gagal!${NC}"; exit 1
  fi
  cd ..
}

echo -e "${CYAN}  RDNS V2.9 PRO - AUTO INSTALLER${NC}"
echo -e "${BLUE}  ================================${NC}"
echo ""
detect_env
echo -e "${CYAN}[*] Mengecek dependensi...${NC}"
install_pkg golang go
install_pkg openssh ssh
install_pkg curl curl
verify_md5
extract_archive
build_binary
echo ""
echo -e "${GREEN}[OK] SELESAI!${NC}"
echo -e "${CYAN}  Jalankan : cd $INSTALL_DIR && ./$BINARY${NC}"
echo -e "${CYAN}  Panel    : http://localhost:8080/?admin=rdns${NC}"
