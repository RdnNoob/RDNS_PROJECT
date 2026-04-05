#!/bin/bash
set -e
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BLUE='\033[0;34m'
NC='\033[0m'

ARCHIVE="rdns_project.tar.gz"
EXPECTED_MD5="ad2885b61eb040a4bcc5c58079b0b43f"
INSTALL_DIR="rdns"
BINARY="rdns_master"

detect_env(){
  if [ -d "/data/data/com.termux/files/usr/bin" ]; then
    ENV="termux"
    echo -e "${CYAN}[*] Environment: Termux${NC}"
  elif command -v apt-get &>/dev/null; then
    ENV="debian"
    echo -e "${CYAN}[*] Environment: Debian/Ubuntu${NC}"
  elif command -v yum &>/dev/null; then
    ENV="redhat"
    echo -e "${CYAN}[*] Environment: RedHat/CentOS${NC}"
  else
    ENV="other"
    echo -e "${YELLOW}[*] Environment: Unknown${NC}"
  fi
}

install_pkg(){
  local P=$1 C=${2:-$1}
  command -v "$C" &>/dev/null && { echo -e "${GREEN}[+] $P sudah ada${NC}"; return; }
  echo -e "${YELLOW}[*] Install $P...${NC}"
