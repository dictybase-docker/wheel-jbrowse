#!/bin/bash
if [ "${SEQDB_PASS+defined} -a ${SEQDB_USER+defined}" ]
then
        gosu postgres postgres --single -E <<-EOSQL
            CREATE ROLE $SEQDB_USER WITH CREATEDB LOGIN ENCRYPTED PASSWORD '$SEQDB_PASS';
            CREATE DATABASE discoideum OWNER $SEQDB_USER;
            CREATE DATABASE purpureum OWNER $SEQDB_USER;
            CREATE DATABASE fasciculatum OWNER $SEQDB_USER;
            CREATE DATABASE pallidum OWNER $SEQDB_USER;
EOSQL
fi
