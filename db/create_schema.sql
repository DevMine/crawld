--
-- PostgreSQL database dump
--

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET search_path = public, pg_catalog;

SET default_with_oids = false;

--
-- Name: gh_organizations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE gh_organizations (
    id bigint NOT NULL,
    github_id bigint NOT NULL,
    login character varying NOT NULL,
    avatar_url character varying,
    html_url character varying,
    name character varying,
    company character varying,
    blog character varying,
    location character varying,
    email character varying,
    collaborators_count integer,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: gh_organizations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE gh_organizations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: gh_organizations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE gh_organizations_id_seq OWNED BY gh_organizations.id;


--
-- Name: gh_repositories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE gh_repositories (
    id bigint NOT NULL,
    repository_id bigint NOT NULL,
    github_id bigint NOT NULL,
    full_name character varying,
    description character varying,
    homepage character varying,
    fork boolean,
    default_branch character varying,
    master_branch character varying,
    html_url character varying,
    forks_count integer,
    open_issues_count integer,
    stargazers_count integer,
    subscribers_count integer,
    watchers_count integer,
    size_in_kb integer,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    pushed_at timestamp with time zone
);


--
-- Name: COLUMN gh_repositories.size_in_kb; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN gh_repositories.size_in_kb IS 'Size of a bare git repository, in kilobytes.';


--
-- Name: gh_repositories_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE gh_repositories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: gh_repositories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE gh_repositories_id_seq OWNED BY gh_repositories.id;


--
-- Name: gh_users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE gh_users (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    github_id bigint NOT NULL,
    login character varying NOT NULL,
    bio text,
    blog character varying,
    company character varying,
    email character varying,
    hireable boolean,
    location character varying,
    avatar_url character varying,
    html_url character varying,
    followers_count integer,
    following_count integer,
    collaborators_count integer,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: gh_users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE gh_users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: gh_users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE gh_users_id_seq OWNED BY gh_users.id;


--
-- Name: gh_users_organizations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE gh_users_organizations (
    gh_user_id bigint NOT NULL,
    gh_organization_id bigint NOT NULL
);


--
-- Name: repositories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE repositories (
    id bigint NOT NULL,
    name character varying NOT NULL,
    primary_language character varying NOT NULL,
    clone_url character varying NOT NULL,
    clone_path character varying NOT NULL,
    vcs character varying NOT NULL
);


--
-- Name: repositories_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE repositories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: repositories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE repositories_id_seq OWNED BY repositories.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE users (
    id bigint NOT NULL,
    username character varying NOT NULL,
    name character varying,
    email character varying
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE users_id_seq OWNED BY users.id;


--
-- Name: users_repositories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE users_repositories (
    user_id bigint NOT NULL,
    repository_id bigint NOT NULL
);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY gh_organizations ALTER COLUMN id SET DEFAULT nextval('gh_organizations_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY gh_repositories ALTER COLUMN id SET DEFAULT nextval('gh_repositories_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY gh_users ALTER COLUMN id SET DEFAULT nextval('gh_users_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY repositories ALTER COLUMN id SET DEFAULT nextval('repositories_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY users ALTER COLUMN id SET DEFAULT nextval('users_id_seq'::regclass);


--
-- Name: gh_organizations_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY gh_organizations
    ADD CONSTRAINT gh_organizations_pk PRIMARY KEY (id);


--
-- Name: gh_organizations_unique_login; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY gh_organizations
    ADD CONSTRAINT gh_organizations_unique_login UNIQUE (login);


--
-- Name: gh_repositories_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY gh_repositories
    ADD CONSTRAINT gh_repositories_pk PRIMARY KEY (id);


--
-- Name: gh_users_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY gh_users
    ADD CONSTRAINT gh_users_pk PRIMARY KEY (id);


--
-- Name: gh_users_unique_login; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY gh_users
    ADD CONSTRAINT gh_users_unique_login UNIQUE (login);


--
-- Name: repositories_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY repositories
    ADD CONSTRAINT repositories_pk PRIMARY KEY (id);


--
-- Name: repositories_unique_clone_path; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY repositories
    ADD CONSTRAINT repositories_unique_clone_path UNIQUE (clone_path);


--
-- Name: repositories_unique_clone_url; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY repositories
    ADD CONSTRAINT repositories_unique_clone_url UNIQUE (clone_url);


--
-- Name: users_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY users
    ADD CONSTRAINT users_pk PRIMARY KEY (id);


--
-- Name: gh_repositories_fk_repositories; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY gh_repositories
    ADD CONSTRAINT gh_repositories_fk_repositories FOREIGN KEY (repository_id) REFERENCES repositories(id);


--
-- Name: gh_users_fk_users; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY gh_users
    ADD CONSTRAINT gh_users_fk_users FOREIGN KEY (user_id) REFERENCES users(id);


--
-- Name: gh_users_organizations_fk_organization; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY gh_users_organizations
    ADD CONSTRAINT gh_users_organizations_fk_organization FOREIGN KEY (gh_organization_id) REFERENCES gh_organizations(id);


--
-- Name: gh_users_organizations_fk_users; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY gh_users_organizations
    ADD CONSTRAINT gh_users_organizations_fk_users FOREIGN KEY (gh_user_id) REFERENCES gh_users(id);


--
-- Name: users_repositories_fk_repository; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY users_repositories
    ADD CONSTRAINT users_repositories_fk_repository FOREIGN KEY (repository_id) REFERENCES repositories(id);


--
-- Name: users_repositories_fk_users; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY users_repositories
    ADD CONSTRAINT users_repositories_fk_users FOREIGN KEY (user_id) REFERENCES users(id);


--
-- PostgreSQL database dump complete
--

