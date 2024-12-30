import os

import yalies
from dotenv import load_dotenv

load_dotenv()


def fetch_data():
    api = yalies.API(
        os.environ.get("YALIES_TOKEN") or input("Insert Yalies API token: ")
    )

    data = api.people(
        query="", filters={"school_code": ["YC"], "year": [2025, 2026, 2027, 2028]}
    )
    return data


def generate_sql(data, table_name, output_file):
    with open(output_file, "w") as sql_file:
        sql_file.write(
            f"-- SQL script to insert data into the '{table_name}' table\n\n"
        )

        for person in data:
            fullname = f"{person.first_name} {person.last_name}".strip(". ").replace(
                "'", "''"
            )

            email = str(person.email).replace("'", "''")
            if not email or email.lower() == "none":
                continue

            year = int(person.year) if person.year else "NULL"

            insert_statement = f"INSERT INTO {table_name} (name, email, graduating_year) VALUES ('{fullname}', '{email}', {year});\n"
            sql_file.write(insert_statement)

    print(f"SQL file '{output_file}' generated successfully.")


def main():
    table_name = "users"
    output_file = "insert_users.sql"

    data = fetch_data()

    generate_sql(data, table_name, output_file)


if __name__ == "__main__":
    main()
