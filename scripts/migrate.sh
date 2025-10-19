#!/bin/sh

set -e

# Variables
DB_URL="postgresql://postgres:postgres@localhost:5432/observ-db?sslmode=disable"
MIGRATIONS_PATH="./migrations"

# Colores
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

usage() {
    echo "Usage: $0 {up|down|create|version|force} [args]"
    echo ""
    echo "Commands:"
    echo "  up [N]           - Apply all or N migrations"
    echo "  down [N]         - Rollback all or N migrations"
    echo "  create NAME      - Create new migration"
    echo "  version          - Show current version"
    echo "  force VERSION    - Force version"
    echo ""
    exit 1
}

if ! command -v migrate &> /dev/null; then
    echo -e "${RED}Error: 'migrate' not found${NC}"
    echo "Install: brew install golang-migrate"
    exit 1
fi

if [ $# -eq 0 ]; then
    usage
fi

COMMAND=$1

case $COMMAND in
    up)
        echo -e "${GREEN}Running migrations UP...${NC}"
        if [ -n "$2" ]; then
            migrate -path "$MIGRATIONS_PATH" -database "$DB_URL" up "$2"
        else
            migrate -path "$MIGRATIONS_PATH" -database "$DB_URL" up
        fi
        echo -e "${GREEN}✓ Migrations applied${NC}"
        ;;
    
    down)
        echo -e "${YELLOW}Running migrations DOWN...${NC}"
        if [ -n "$2" ]; then
            migrate -path "$MIGRATIONS_PATH" -database "$DB_URL" down "$2"
        else
            echo -e "${RED}Warning: This will rollback ALL migrations!${NC}"
            read -p "Are you sure? (yes/no): " confirm
            if [ "$confirm" = "yes" ]; then
                migrate -path "$MIGRATIONS_PATH" -database "$DB_URL" down
            else
                echo "Aborted"
                exit 0
            fi
        fi
        echo -e "${YELLOW}✓ Rollback complete${NC}"
        ;;
    
    create)
        if [ -z "$2" ]; then
            echo -e "${RED}Error: Migration name required${NC}"
            exit 1
        fi
        echo -e "${GREEN}Creating migration: $2${NC}"
        migrate create -ext sql -dir "$MIGRATIONS_PATH" -seq "$2"
        echo -e "${GREEN}✓ Migration files created${NC}"
        ;;
    
    version)
        echo -e "${GREEN}Current version:${NC}"
        migrate -path "$MIGRATIONS_PATH" -database "$DB_URL" version
        ;;
    
    force)
        if [ -z "$2" ]; then
            echo -e "${RED}Error: Version required${NC}"
            exit 1
        fi
        migrate -path "$MIGRATIONS_PATH" -database "$DB_URL" force "$2"
        echo -e "${YELLOW}✓ Version forced to $2${NC}"
        ;;
    
    *)
        echo -e "${RED}Unknown command: $COMMAND${NC}"
        usage
        ;;
esac