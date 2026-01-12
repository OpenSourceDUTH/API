#!/bin/bash

# --- Styles ---
NORMAL="\033[0m"
BOLD="\033[1;37m"
GREEN="\033[32m"
RED="\033[31m"
YELLOW="\033[33m"
CYAN="\033[36m"     

function echo_logo {
    echo -e "${CYAN}${BOLD}$1${NORMAL}"
}

function echo_info {
    echo -e "${BOLD}${CYAN}==>${NORMAL} $1"
} 

function echo_step {
    echo -e "\n${YELLOW}--- STEP $1 ---${NORMAL}"
}

function echo_success {
    echo -e "${GREEN}SUCCESS:${NORMAL} $1"
}

function echo_error {
    echo -e "${RED}ERROR:${NORMAL} $1" >&2
}

# --- Banner ---
echo_logo " _____                  _____                         ______ _   _ _____ _   _ "
echo_logo "|  _  |                /  ___|                        |  _  \ | | |_   _| | | |"
echo_logo "| | | |_ __   ___ _ __ \ \`--.  ___  _   _ _ __ ___ ___| | | | | | | | | | |_| |"
echo_logo "| | | | '_ \ / _ \ '_ \ \`--. \/ _ \| | | | '__/ __/ _ \ | | | | | | | | |  _  |"
echo_logo "\ \_/ / |_) |  __/ | | /\__/ / (_) | |_| | | | (_|  __/ |/ /| |_| | | | | | | |"
echo_logo " \___/| .__/ \___|_| |_\____/ \___/ \__,_|_|  \___\___|___/  \___/  \_/ \_| |_/"
echo_logo "      | |                                                                      "
echo_logo "      |_|                                                                      "

echo ""
echo_info "Welcome to the OpenSourceDUTH Environment Setup Tool"
echo "This script will help you automate the setup of this project."
echo "This tool is for development and early testing environments, for production deployments, please refer to the official documentation or ensure that this script meets your security and configuration standards. You have been warned!"
echo ""
echo "It will:"
echo " 1. "
echo " 2. "
echo " 3. "

echo_error "This project is not yet ready for deployment. This tool does not work."

echo -e "\n${BOLD}${YELLOW}Press any key to get started, or Ctrl+C to exit...${NORMAL}"


read -n 1 -s

# --- License ---
# This project is the monolithic backend API for the OpenSourceDUTH team. Access to open data compiled and provided by the OpenSourceDUTH University Team.
# API Copyright (C) 2025 OpenSourceDUTH
#     This program is free software: you can redistribute it and/or modify
#     it under the terms of the GNU General Public License as published by
#     the Free Software Foundation, either version 3 of the License, or
#     (at your option) any later version.

#     This program is distributed in the hope that it will be useful,
#     but WITHOUT ANY WARRANTY; without even the implied warranty of
#     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
#     GNU General Public License for more details.

#     You should have received a copy of the GNU General Public License
#     along with this program.  If not, see <https://www.gnu.org/licenses/>.