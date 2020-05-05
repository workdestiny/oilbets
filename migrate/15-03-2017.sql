--
-- PostgreSQL database dump
--

-- Dumped from database version 9.5.12
-- Dumped by pg_dump version 9.5.12

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner:
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner:
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: category; Type: TABLE; Schema: public; Owner: postgre_dev
--

CREATE TABLE public.category (
    id character varying NOT NULL,
    code character varying NOT NULL,
    count integer NOT NULL,
    created_at character varying NOT NULL,
    images json,
    name json,
    updated_at character varying,
    verify boolean
);


ALTER TABLE public.category OWNER TO postgre_dev;

--
-- Name: comasdasd_id_seq; Type: SEQUENCE; Schema: public; Owner: postgre_dev
--

CREATE SEQUENCE public.comasdasd_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.comasdasd_id_seq OWNER TO postgre_dev;

--
-- Name: comment_id_seq; Type: SEQUENCE; Schema: public; Owner: postgre_dev
--

CREATE SEQUENCE public.comment_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.comment_id_seq OWNER TO postgre_dev;

--
-- Name: comment; Type: TABLE; Schema: public; Owner: postgre_dev
--

CREATE TABLE public.comment (
    id integer DEFAULT nextval('public.comment_id_seq'::regclass) NOT NULL,
    post_id integer NOT NULL,
    owner_id character varying(40) NOT NULL,
    owner_type integer NOT NULL,
    text text NOT NULL,
    created_at integer NOT NULL,
    updated_at integer NOT NULL,
    status boolean NOT NULL,
    used boolean NOT NULL
);


ALTER TABLE public.comment OWNER TO postgre_dev;

--
-- Name: follow_gap; Type: TABLE; Schema: public; Owner: postgre_dev
--

CREATE TABLE public.follow_gap (
    id character varying NOT NULL,
    created_at character varying NOT NULL,
    owner_id character varying NOT NULL,
    gap_id character varying NOT NULL,
    status boolean NOT NULL,
    used boolean NOT NULL,
    updated_at character varying NOT NULL
);


ALTER TABLE public.follow_gap OWNER TO postgre_dev;

--
-- Name: follow_topic; Type: TABLE; Schema: public; Owner: postgre_dev
--

CREATE TABLE public.follow_topic (
    id character varying NOT NULL,
    created_at character varying NOT NULL,
    owner_id character varying NOT NULL,
    status boolean NOT NULL,
    topic_id character varying NOT NULL,
    used boolean NOT NULL,
    updated_at character varying
);


ALTER TABLE public.follow_topic OWNER TO postgre_dev;

--
-- Name: gap; Type: TABLE; Schema: public; Owner: postgre_dev
--

CREATE TABLE public.gap (
    id character varying NOT NULL,
    bio text,
    cat_id character varying NOT NULL,
    contact json NOT NULL,
    count json NOT NULL,
    cover json NOT NULL,
    display json NOT NULL,
    name json NOT NULL,
    status json NOT NULL,
    topic_id character varying NOT NULL,
    used boolean NOT NULL,
    updated_at character varying NOT NULL,
    user_id character varying NOT NULL,
    username json NOT NULL,
    verify json NOT NULL,
    created_at character varying NOT NULL
);


ALTER TABLE public.gap OWNER TO postgre_dev;

--
-- Name: gap_audit; Type: TABLE; Schema: public; Owner: postgre_dev
--

CREATE TABLE public.gap_audit (
    id character varying NOT NULL,
    created_at character varying NOT NULL,
    gap_id character varying NOT NULL,
    snapshot json NOT NULL,
    type character(10) NOT NULL,
    updated_at character varying NOT NULL
);


ALTER TABLE public.gap_audit OWNER TO postgre_dev;

--
-- Name: gap_audit_id_seq; Type: SEQUENCE; Schema: public; Owner: postgre_dev
--

CREATE SEQUENCE public.gap_audit_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.gap_audit_id_seq OWNER TO postgre_dev;

--
-- Name: gap_audit_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgre_dev
--

ALTER SEQUENCE public.gap_audit_id_seq OWNED BY public.gap_audit.id;


--
-- Name: gap_view; Type: TABLE; Schema: public; Owner: postgre_dev
--

CREATE TABLE public.gap_view (
    id character varying NOT NULL,
    created_at character varying,
    owner_id character varying,
    gap_id character varying,
    updated_at character varying
);


ALTER TABLE public.gap_view OWNER TO postgre_dev;

--
-- Name: gap_view_id_seq; Type: SEQUENCE; Schema: public; Owner: postgre_dev
--

CREATE SEQUENCE public.gap_view_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.gap_view_id_seq OWNER TO postgre_dev;

--
-- Name: gap_view_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgre_dev
--

ALTER SEQUENCE public.gap_view_id_seq OWNED BY public.gap_view.id;


--
-- Name: imageinbucket_id_seq; Type: SEQUENCE; Schema: public; Owner: postgre_dev
--

CREATE SEQUENCE public.imageinbucket_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.imageinbucket_id_seq OWNER TO postgre_dev;

--
-- Name: imageinbucket; Type: TABLE; Schema: public; Owner: postgre_dev
--

CREATE TABLE public.imageinbucket (
    id bigint DEFAULT nextval('public.imageinbucket_id_seq'::regclass) NOT NULL,
    owner_id character varying(40) NOT NULL,
    image text NOT NULL,
    height integer DEFAULT 0 NOT NULL,
    width integer DEFAULT 0 NOT NULL,
    status boolean NOT NULL,
    created_at integer NOT NULL
);


ALTER TABLE public.imageinbucket OWNER TO postgre_dev;

--
-- Name: likepost_id_seq; Type: SEQUENCE; Schema: public; Owner: postgre_dev
--

CREATE SEQUENCE public.likepost_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.likepost_id_seq OWNER TO postgre_dev;

--
-- Name: likepost; Type: TABLE; Schema: public; Owner: postgre_dev
--

CREATE TABLE public.likepost (
    id integer DEFAULT nextval('public.likepost_id_seq'::regclass) NOT NULL,
    post_id integer NOT NULL,
    owner_id character varying(40) NOT NULL,
    created_at integer NOT NULL,
    updated_at integer NOT NULL,
    status boolean NOT NULL,
    used boolean NOT NULL
);


ALTER TABLE public.likepost OWNER TO postgre_dev;

--
-- Name: notification_id_seq; Type: SEQUENCE; Schema: public; Owner: postgre_dev
--

CREATE SEQUENCE public.notification_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.notification_id_seq OWNER TO postgre_dev;

--
-- Name: notification; Type: TABLE; Schema: public; Owner: postgre_dev
--

CREATE TABLE public.notification (
    id integer DEFAULT nextval('public.notification_id_seq'::regclass) NOT NULL,
    owner_id character varying(40) NOT NULL,
    post_id integer NOT NULL,
    gap_id character varying(40) NOT NULL,
    created_at integer NOT NULL,
    updated_at integer NOT NULL,
    title text,
    count integer NOT NULL,
    type integer NOT NULL,
    type_post integer NOT NULL,
    read boolean NOT NULL,
    status boolean NOT NULL,
    firebase boolean NOT NULL
);


ALTER TABLE public.notification OWNER TO postgre_dev;

--
-- Name: post_id_seq; Type: SEQUENCE; Schema: public; Owner: postgre_dev
--

CREATE SEQUENCE public.post_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.post_id_seq OWNER TO postgre_dev;

--
-- Name: post; Type: TABLE; Schema: public; Owner: postgre_dev
--

CREATE TABLE public.post (
    id integer DEFAULT nextval('public.post_id_seq'::regclass) NOT NULL,
    code character varying(510) NOT NULL,
    owner_id character varying(40) NOT NULL,
    owner_type integer NOT NULL,
    title text NOT NULL,
    description text NOT NULL,
    image_url text NOT NULL,
    height integer DEFAULT 0 NOT NULL,
    width integer DEFAULT 0 NOT NULL,
    link text NOT NULL,
    link_description text NOT NULL,
    created_at integer NOT NULL,
    updated_at integer NOT NULL,
    type integer NOT NULL,
    province character varying(10) NOT NULL,
    like_count integer DEFAULT 0 NOT NULL,
    comment_count integer DEFAULT 0 NOT NULL,
    share_count integer DEFAULT 0 NOT NULL,
    view_count integer DEFAULT 0 NOT NULL,
    status boolean NOT NULL,
    verify boolean NOT NULL,
    used boolean NOT NULL
);


ALTER TABLE public.post OWNER TO postgre_dev;

--
-- Name: statistic_token; Type: TABLE; Schema: public; Owner: postgre_dev
--

CREATE TABLE public.statistic_token (
    id character varying NOT NULL,
    created_at character varying NOT NULL,
    location json NOT NULL,
    refresh_token character varying NOT NULL,
    signin_type integer NOT NULL,
    signout_at character varying NOT NULL,
    updated_at character varying NOT NULL,
    status boolean NOT NULL,
    user_agent text,
    user_id character varying NOT NULL
);


ALTER TABLE public.statistic_token OWNER TO postgre_dev;

--
-- Name: statistic_token_id_seq; Type: SEQUENCE; Schema: public; Owner: postgre_dev
--

CREATE SEQUENCE public.statistic_token_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.statistic_token_id_seq OWNER TO postgre_dev;

--
-- Name: statistic_token_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgre_dev
--

ALTER SEQUENCE public.statistic_token_id_seq OWNED BY public.statistic_token.id;


--
-- Name: subnotification_id_seq; Type: SEQUENCE; Schema: public; Owner: postgre_dev
--

CREATE SEQUENCE public.subnotification_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.subnotification_id_seq OWNER TO postgre_dev;

--
-- Name: subnotification; Type: TABLE; Schema: public; Owner: postgre_dev
--

CREATE TABLE public.subnotification (
    id integer DEFAULT nextval('public.subnotification_id_seq'::regclass) NOT NULL,
    notification_id integer NOT NULL,
    tag_id integer NOT NULL,
    tag_type integer NOT NULL,
    owner_id character varying(40) NOT NULL,
    status boolean NOT NULL,
    created_at integer NOT NULL,
    updated_at integer NOT NULL
);


ALTER TABLE public.subnotification OWNER TO postgre_dev;

--
-- Name: tagtopic_id_seq; Type: SEQUENCE; Schema: public; Owner: postgre_dev
--

CREATE SEQUENCE public.tagtopic_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.tagtopic_id_seq OWNER TO postgre_dev;

--
-- Name: tagtopic; Type: TABLE; Schema: public; Owner: postgre_dev
--

CREATE TABLE public.tagtopic (
    id integer DEFAULT nextval('public.tagtopic_id_seq'::regclass) NOT NULL,
    post_id integer NOT NULL,
    topic_id character varying(40) NOT NULL,
    created_at integer NOT NULL
);


ALTER TABLE public.tagtopic OWNER TO postgre_dev;

--
-- Name: topic; Type: TABLE; Schema: public; Owner: postgre_dev
--

CREATE TABLE public.topic (
    id character varying NOT NULL,
    cat_id character varying NOT NULL,
    code character varying NOT NULL,
    count integer NOT NULL,
    created_at character varying NOT NULL,
    images json,
    name json NOT NULL,
    updated_at character varying,
    used_count integer NOT NULL,
    verify boolean NOT NULL
);


ALTER TABLE public.topic OWNER TO postgre_dev;

--
-- Name: user; Type: TABLE; Schema: public; Owner: postgre_dev
--

CREATE TABLE public."user" (
    id character varying NOT NULL,
    about_me json,
    birthdate character varying NOT NULL,
    contact json NOT NULL,
    count json NOT NULL,
    created_at character varying NOT NULL,
    display json NOT NULL,
    email json NOT NULL,
    firstname character varying NOT NULL,
    gender character(8) NOT NULL,
    gap boolean NOT NULL,
    lastname character varying NOT NULL,
    role integer NOT NULL,
    status json NOT NULL,
    used boolean NOT NULL,
    updated_at character varying NOT NULL,
    username json NOT NULL,
    verify json NOT NULL
);


ALTER TABLE public."user" OWNER TO postgre_dev;

--
-- Name: user_audit; Type: TABLE; Schema: public; Owner: postgre_dev
--

CREATE TABLE users_audit (
    id character varying NOT NULL,
    created_at character varying NOT NULL,
    user_id character varying NOT NULL,
    snapshot json NOT NULL,
    type character(10) NOT NULL,
    updated_at character varying NOT NULL
);


ALTER TABLE users_audit OWNER TO postgre_dev;

--
-- Name: user_provider; Type: TABLE; Schema: public; Owner: postgre_dev
--

CREATE TABLE user_provider (
    id character varying NOT NULL,
    code_social_signin text,
    count integer,
    created_at character varying,
    provider json,
    "time" character varying,
    update_email json,
    updated_at character varying
);


ALTER TABLE user_provider OWNER TO postgre_dev;

--
-- Name: user_provider_id_seq; Type: SEQUENCE; Schema: public; Owner: postgre_dev
--

CREATE SEQUENCE user_provider_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE user_provider_id_seq OWNER TO postgre_dev;

--
-- Name: user_provider_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgre_dev
--

ALTER SEQUENCE user_provider_id_seq OWNED BY user_provider.id;


--
-- Name: view; Type: TABLE; Schema: public; Owner: postgre_dev
--

CREATE TABLE public.view (
    id character varying(510) NOT NULL,
    post_id integer NOT NULL,
    owner_id character varying(40) NOT NULL,
    created_at character varying NOT NULL
);


ALTER TABLE public.view OWNER TO postgre_dev;

--
-- Name: category_pkey; Type: CONSTRAINT; Schema: public; Owner: postgre_dev
--

ALTER TABLE ONLY public.category
    ADD CONSTRAINT category_pkey PRIMARY KEY (id);


--
-- Name: comment_pkey; Type: CONSTRAINT; Schema: public; Owner: postgre_dev
--

ALTER TABLE ONLY public.comment
    ADD CONSTRAINT comment_pkey PRIMARY KEY (id);


--
-- Name: follow_gap_pkey; Type: CONSTRAINT; Schema: public; Owner: postgre_dev
--

ALTER TABLE ONLY public.follow_gap
    ADD CONSTRAINT follow_gap_pkey PRIMARY KEY (id);


--
-- Name: follow_topic_pkey; Type: CONSTRAINT; Schema: public; Owner: postgre_dev
--

ALTER TABLE ONLY public.follow_topic
    ADD CONSTRAINT follow_topic_pkey PRIMARY KEY (id);


--
-- Name: gap_audit_pkey; Type: CONSTRAINT; Schema: public; Owner: postgre_dev
--

ALTER TABLE ONLY public.gap_audit
    ADD CONSTRAINT gap_audit_pkey PRIMARY KEY (id);


--
-- Name: gap_pkey; Type: CONSTRAINT; Schema: public; Owner: postgre_dev
--

ALTER TABLE ONLY public.gap
    ADD CONSTRAINT gap_pkey PRIMARY KEY (id);


--
-- Name: gap_view_pkey; Type: CONSTRAINT; Schema: public; Owner: postgre_dev
--

ALTER TABLE ONLY public.gap_view
    ADD CONSTRAINT gap_view_pkey PRIMARY KEY (id);


--
-- Name: imageinbucket_pkey; Type: CONSTRAINT; Schema: public; Owner: postgre_dev
--

ALTER TABLE ONLY public.imageinbucket
    ADD CONSTRAINT imageinbucket_pkey PRIMARY KEY (id);


--
-- Name: likepost_pkey; Type: CONSTRAINT; Schema: public; Owner: postgre_dev
--

ALTER TABLE ONLY public.likepost
    ADD CONSTRAINT likepost_pkey PRIMARY KEY (id);


--
-- Name: notification_pkey; Type: CONSTRAINT; Schema: public; Owner: postgre_dev
--

ALTER TABLE ONLY public.notification
    ADD CONSTRAINT notification_pkey PRIMARY KEY (id);


--
-- Name: post_pkey; Type: CONSTRAINT; Schema: public; Owner: postgre_dev
--

ALTER TABLE ONLY public.post
    ADD CONSTRAINT post_pkey PRIMARY KEY (id);


--
-- Name: statistic_token_pkey; Type: CONSTRAINT; Schema: public; Owner: postgre_dev
--

ALTER TABLE ONLY public.statistic_token
    ADD CONSTRAINT statistic_token_pkey PRIMARY KEY (id);


--
-- Name: subnotification_pkey; Type: CONSTRAINT; Schema: public; Owner: postgre_dev
--

ALTER TABLE ONLY public.subnotification
    ADD CONSTRAINT subnotification_pkey PRIMARY KEY (id);


--
-- Name: tagtopic_pkey; Type: CONSTRAINT; Schema: public; Owner: postgre_dev
--

ALTER TABLE ONLY public.tagtopic
    ADD CONSTRAINT tagtopic_pkey PRIMARY KEY (id);


--
-- Name: topic_pkey; Type: CONSTRAINT; Schema: public; Owner: postgre_dev
--

ALTER TABLE ONLY public.topic
    ADD CONSTRAINT topic_pkey PRIMARY KEY (id);


--
-- Name: untitled_table_pkey; Type: CONSTRAINT; Schema: public; Owner: postgre_dev
--

ALTER TABLE ONLY users_audit
    ADD CONSTRAINT untitled_table_pkey PRIMARY KEY (id);


--
-- Name: user_pkey; Type: CONSTRAINT; Schema: public; Owner: postgre_dev
--

ALTER TABLE ONLY public."user"
    ADD CONSTRAINT user_pkey PRIMARY KEY (id);


--
-- Name: user_provider_pkey; Type: CONSTRAINT; Schema: public; Owner: postgre_dev
--

ALTER TABLE ONLY user_provider
    ADD CONSTRAINT user_provider_pkey PRIMARY KEY (id);


--
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: postgres
--

REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM postgres;
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- PostgreSQL database dump complete
--

