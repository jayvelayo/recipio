#!/usr/bin/env python3
"""Delete a Recipio user and all their data by email address."""

import argparse
import os
import sqlite3
import sys

def default_db_path():
    system = os.uname().sysname
    if system == "Darwin":
        cache = os.path.expanduser("~/Library/Caches")
    else:
        cache = os.environ.get("XDG_CACHE_HOME", os.path.expanduser("~/.cache"))
    return os.path.join(cache, "recipio", "recipes.db")

def delete_user(db_path: str, email: str, force: bool) -> None:
    if not os.path.exists(db_path):
        sys.exit(f"Error: database not found at {db_path}")

    con = sqlite3.connect(db_path)
    con.row_factory = sqlite3.Row

    try:
        user = con.execute("SELECT id, name, email, created FROM users WHERE email = ?", (email,)).fetchone()
        if user is None:
            sys.exit(f"No user found with email: {email}")

        user_id = user["id"]
        print(f"User found:")
        print(f"  ID:      {user_id}")
        print(f"  Name:    {user['name']}")
        print(f"  Email:   {user['email']}")
        print(f"  Created: {user['created']}")

        counts = {
            "recipes":     con.execute("SELECT COUNT(*) FROM recipes WHERE user_id = ?", (user_id,)).fetchone()[0],
            "meal plans":  con.execute("SELECT COUNT(*) FROM meal_plan WHERE user_id = ?", (user_id,)).fetchone()[0],
            "grocery lists": con.execute("SELECT COUNT(*) FROM grocery_lists WHERE user_id = ?", (user_id,)).fetchone()[0],
            "sessions":    con.execute("SELECT COUNT(*) FROM sessions WHERE user_id = ?", (user_id,)).fetchone()[0],
        }
        print("\nData that will be permanently deleted:")
        for label, count in counts.items():
            print(f"  {count} {label}")

        if not force:
            answer = input("\nDelete this user and all their data? [y/N] ").strip().lower()
            if answer != "y":
                print("Aborted.")
                return

        with con:
            # Child records first to satisfy foreign key constraints
            recipe_ids = [r[0] for r in con.execute("SELECT id FROM recipes WHERE user_id = ?", (user_id,)).fetchall()]
            if recipe_ids:
                con.execute(f"DELETE FROM ingredients WHERE recipe_id IN ({','.join('?' * len(recipe_ids))})", recipe_ids)

            meal_plan_ids = [r[0] for r in con.execute("SELECT id FROM meal_plan WHERE user_id = ?", (user_id,)).fetchall()]
            if meal_plan_ids:
                con.execute(f"DELETE FROM meal_plan_recipes WHERE meal_plan_id IN ({','.join('?' * len(meal_plan_ids))})", meal_plan_ids)

            grocery_list_ids = [r[0] for r in con.execute("SELECT id FROM grocery_lists WHERE user_id = ?", (user_id,)).fetchall()]
            if grocery_list_ids:
                con.execute(f"DELETE FROM grocery_list_items WHERE grocery_list_id IN ({','.join('?' * len(grocery_list_ids))})", grocery_list_ids)

            con.execute("DELETE FROM email_verifications WHERE user_id = ?", (user_id,))
            con.execute("DELETE FROM sessions WHERE user_id = ?", (user_id,))
            con.execute("DELETE FROM credentials WHERE user_id = ?", (user_id,))
            con.execute("DELETE FROM oauth WHERE user_id = ?", (user_id,))
            con.execute("DELETE FROM grocery_lists WHERE user_id = ?", (user_id,))
            con.execute("DELETE FROM meal_plan WHERE user_id = ?", (user_id,))
            con.execute("DELETE FROM recipes WHERE user_id = ?", (user_id,))
            con.execute("DELETE FROM users WHERE id = ?", (user_id,))

        print(f"\nDeleted user {email}.")
    finally:
        con.close()

def main():
    parser = argparse.ArgumentParser(description="Delete a Recipio user by email.")
    parser.add_argument("email", help="Email address of the user to delete")
    parser.add_argument("--db", default=default_db_path(), help="Path to the SQLite database")
    parser.add_argument("-f", "--force", action="store_true", help="Skip confirmation prompt")
    args = parser.parse_args()

    delete_user(args.db, args.email, args.force)

if __name__ == "__main__":
    main()
