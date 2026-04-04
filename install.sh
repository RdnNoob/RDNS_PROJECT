#!/bin/bash

# --- KONFIGURASI WARNA ---
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${PURPLE}==========================================${NC}"
echo -e "${CYAN}      RDNS V2.9 PRO AUTO-INSTALLER       ${NC}"
echo -e "${PURPLE}==========================================${NC}"

# 1. Deteksi Package Manager
if [ -d "/data/data/com.termux/files/usr/bin" ]; then
    PM="pkg"
else
    PM="sudo apt-get"
fi

# 2. Update & Install Dependensi
echo -e "${YELLOW}[*] Checking dependencies...${NC}"
$PM update -y
$PM install golang openssh git curl -y

# 3. Setup Go Module
echo -e "${YELLOW}[*] Setting up Go environment...${NC}"
if [ ! -f "go.mod" ]; then
    go mod init rdns_suite
fi
go mod tidy

# 4. Build Binary
echo -e "${YELLOW}[*] Compiling RDNS Master Engine...${NC}"
go build -o rdns_master main.go ui.go

if [ $? -eq 0 ]; then
    chmod +x rdns_master
    echo -e "${PURPLE}==========================================${NC}"
    echo -e "${CYAN}[+] INSTALASI BERHASIL!${NC}"
    echo -e "${YELLOW}[!] Jalankan dengan: ./rdns_master${NC}"
    echo -e "${PURPLE}==========================================${NC}"
else
    echo -e "\033[0;31m[-] Build failed. Pastikan main.go dan ui.go ada di folder ini.\033[0m"
fi
