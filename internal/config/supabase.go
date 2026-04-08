package config

import (
    "log"
    "os"

    supa "github.com/supabase-community/supabase-go"
)

var SupabaseClient *supa.Client

func ConnectSupabase() {
    url := os.Getenv("SUPABASE_URL")
    key := os.Getenv("SUPABASE_KEY")

    client, err := supa.NewClient(url, key, &supa.ClientOptions{})
    if err != nil {
        log.Fatal("❌ Failed to init Supabase client:", err)
    }

    log.Println("✅ Supabase client initialized")
    SupabaseClient = client
}