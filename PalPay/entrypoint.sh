#!/bin/sh
set -e

if [ ! -f /app/palpay.db ]; then
    sqlite3 /app/palpay.db < /app/init_db.sql
fi
# Avvia l'applicazione
chmod 666 /app/palpay.db

exec ./palpay