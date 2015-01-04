# Database schema creation script

The database in use is PostgresSQL 9.3+.
This script creates the tables required for the crawler, ie:

 * **users**: table to store general users information.
 * **repositories**: table to store general repositories information.
 * **gh\_users**: table to store information about GitHub users.
 * **gh\_repositories**: table to store information about GitHub repositories.
 * **gh\_organizations**: table to store information about GitHub organizations.

And 2 relation tables:

 * **users\_repositories**: links users to the repositories they contributed to.
 * **gh\_users\_organizations**: links GitHub users to the GitHub organizations
   they belong to.

You need to create an empty PostgreSQL database, UTF8 encoded and then run:

    psql -U user dbname < create_schema.sql
