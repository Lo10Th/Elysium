-- Elysium Registry Database Schema
-- Compatible with Supabase API v2+
-- Run this in your Supabase SQL Editor (Dashboard > SQL Editor)

-- Enable UUID extension
create extension if not exists "uuid-ossp";

-- ============================================================================
-- USERS (profiles)
-- ============================================================================

-- User profiles table (extends Supabase auth.users)
create table if not exists profiles (
    id uuid references auth.users on delete cascade primary key,
    username text unique not null,
    email text not null,
    avatar_url text,
    bio text,
    created_at timestamp with time zone default timezone('utc'::text, now()) not null,
    updated_at timestamp with time zone default timezone('utc'::text, now()) not null
);

-- Enable RLS
alter table profiles enable row level security;

-- Policies
create policy "Users can view own profile"
    on profiles for select
    using (auth.uid() = id);

create policy "Users can update own profile"
    on profiles for update
    using (auth.uid() = id);

create policy "Anyone can view profiles"
    on profiles for select
    using (true);

-- Trigger to create profile on user signup
create or replace function public.handle_new_user()
returns trigger
language plpgsql
security definer
as $$
begin
    insert into public.profiles (id, username, email)
    values (
        new.id,
        new.raw_user_meta_data->>'username',
        new.email
    );
    return new;
end;
$$;

-- Trigger
drop trigger if exists on_auth_user_created on auth.users;
create trigger on_auth_user_created
    after insert on auth.users
    for each row execute procedure public.handle_new_user();

-- ============================================================================
-- EMBLEMS
-- ============================================================================

-- Emblems table (APIs)
create table if not exists emblems (
    id uuid default uuid_generate_v4() primary key,
    name text unique not null check (name ~ '^[a-z0-9][a-z0-9-]*[a-z0-9]$'),
    description text not null check (char_length(description) >= 10),
    author_id uuid references profiles(id) on delete cascade not null,
    category text,
    tags text[] default '{}',
    license text default 'MIT',
    repository_url text,
    homepage_url text,
    downloads_count integer default 0,
    created_at timestamp with time zone default timezone('utc'::text, now()) not null,
    updated_at timestamp with time zone default timezone('utc'::text, now()) not null
);

-- Enable RLS
alter table emblems enable row level security;

-- Policies
create policy "Anyone can view emblems"
    on emblems for select
    using (true);

create policy "Authors can create emblems"
    on emblems for insert
    with check (auth.uid() = author_id);

create policy "Authors can update own emblems"
    on emblems for update
    using (auth.uid() = author_id);

create policy "Authors can delete own emblems"
    on emblems for delete
    using (auth.uid() = author_id);

-- Indexes
create index idx_emblems_name on emblems(name);
create index idx_emblems_category on emblems(category);
create index idx_emblems_tags on emblems using gin(tags);
create index idx_emblems_author on emblems(author_id);

-- ============================================================================
-- EMBLEM_VERSIONS
-- ============================================================================

-- Emblem versions table
create table if not exists emblem_versions (
    id uuid default uuid_generate_v4() primary key,
    emblem_id uuid references emblems(id) on delete cascade not null,
    version text not null check (version ~ '^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$'),
    yaml_content text not null,
    readme_content text,
    changelog text,
    created_at timestamp with time zone default timezone('utc'::text, now()) not null,
    
    unique(emblem_id, version)
);

-- Enable RLS
alter table emblem_versions enable row level security;

-- Policies
create policy "Anyone can view emblem versions"
    on emblem_versions for select
    using (true);

create policy "Authors can create versions"
    on emblem_versions for insert
    with check (
        exists (
            select 1 from emblems
            where emblems.id = emblem_versions.emblem_id
            and emblems.author_id = auth.uid()
        )
    );

-- Indexes
create index idx_emblem_versions_emblem on emblem_versions(emblem_id);
create index idx_emblem_versions_version on emblem_versions(version);

-- ============================================================================
-- API_KEYS
-- ============================================================================

-- API keys table
create table if not exists api_keys (
    id text primary key default 'key_' || encode(gen_random_bytes(16), 'hex'),
    user_id uuid references profiles(id) on delete cascade not null,
    name text not null check (char_length(name) >= 1 and char_length(name) <= 100),
    key_hash text not null,
    last_used_at timestamp with time zone,
    expires_at timestamp with time zone,
    created_at timestamp with time zone default timezone('utc'::text, now()) not null
);

-- Enable RLS
alter table api_keys enable row level security;

-- Policies
create policy "Users can view own keys"
    on api_keys for select
    using (auth.uid() = user_id);

create policy "Users can create keys"
    on api_keys for insert
    with check (auth.uid() = user_id);

create policy "Users can delete own keys"
    on api_keys for delete
    using (auth.uid() = user_id);

-- Index
create index idx_api_keys_user on api_keys(user_id);
create index idx_api_keys_hash on api_keys(key_hash);

-- ============================================================================
-- EMBLEM_PULLS (Analytics)
-- ============================================================================

-- Track emblem downloads/pulls
create table if not exists emblem_pulls (
    id uuid default uuid_generate_v4() primary key,
    emblem_id uuid references emblems(id) on delete cascade not null,
    version text,
    user_id uuid references profiles(id) on delete set null,
    ip_address inet,
    user_agent text,
    created_at timestamp with time zone default timezone('utc'::text, now()) not null
);

-- Enable RLS
alter table emblem_pulls enable row level security;

-- Policies
create policy "Anyone can insert pulls"
    on emblem_pulls for insert
    with check (true);

create policy "Anyone can view pulls"
    on emblem_pulls for select
    using (true);

-- Index
create index idx_emblem_pulls_emblem on emblem_pulls(emblem_id);
create index idx_emblem_pulls_created on emblem_pulls(created_at);

-- ============================================================================
-- FUNCTIONS & TRIGGERS
-- ============================================================================

-- Function to increment downloads count
create or replace function increment_downloads()
returns trigger
language plpgsql
as $$
begin
    update emblems
    set downloads_count = downloads_count + 1
    where id = new.emblem_id;
    return new;
end;
$$;

-- Trigger
drop trigger if exists on_emblem_pull_created on emblem_pulls;
create trigger on_emblem_pull_created
    after insert on emblem_pulls
    for each row execute procedure increment_downloads();

-- Function to update updated_at timestamp
create or replace function update_updated_at()
returns trigger
language plpgsql
as $$
begin
    new.updated_at = timezone('utc'::text, now());
    return new;
end;
$$;

-- Triggers for updated_at
drop trigger if exists on_profiles_updated on profiles;
create trigger on_profiles_updated
    before update on profiles
    for each row execute procedure update_updated_at();

drop trigger if exists on_emblems_updated on emblems;
create trigger on_emblems_updated
    before update on emblems
    for each row execute procedure update_updated_at();

-- ============================================================================
-- VIEWS
-- ============================================================================

-- View for emblem with latest version and author info
create or replace view emblem_with_details as
select
    e.id,
    e.name,
    e.description,
    e.category,
    e.tags,
    e.license,
    e.repository_url,
    e.homepage_url,
    e.downloads_count,
    e.created_at,
    e.updated_at,
    e.author_id,
    p.username as author_name,
    ev.version as latest_version
from emblems e
left join profiles p on e.author_id = p.id
left join lateral (
    select version
    from emblem_versions
    where emblem_id = e.id
    order by created_at desc
    limit 1
) ev on true;

-- ============================================================================
-- SAMPLE DATA (Optional - for testing)
-- ============================================================================

-- Uncomment to add sample emblem after user signup
-- Note: You'll need to replace 'user-uuid-here' with actual user ID

-- insert into emblems (id, name, description, author_id, category, tags, repository_url, homepage_url)
-- values (
--     uuid_generate_v4(),
--     'clothing-shop',
--     'API for managing an online clothing store with products, categories, and orders',
--     'user-uuid-here',
--     'ecommerce',
--     array['api', 'rest', 'clothing', 'shop'],
--     'https://github.com/example/clothing-shop-api',
--     'https://clothing-shop.example.com'
-- );

-- ============================================================================
-- GRANTS
-- ============================================================================

-- Grant permissions to authenticated users
grant all on all tables in schema public to authenticated;
grant all on all sequences in schema public to authenticated;
grant all on all functions in schema public to authenticated;

-- Grant permissions to anon users (for public read access)
grant select on emblems to anon;
grant select on emblem_versions to anon;
grant select on profiles to anon;