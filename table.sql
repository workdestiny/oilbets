CREATE TABLE category (
    id character varying DEFAULT uuid_generate_v1() PRIMARY KEY,
    code character varying NOT NULL,
    count integer NOT NULL,
    created_at character varying NOT NULL,
    images json,
    name json,
    updated_at character varying,
    verify boolean
);

CREATE UNIQUE INDEX category_pkey ON category(id text_ops);
CREATE INDEX category_created_at_idx ON category(created_at text_ops);
CREATE INDEX category_verify_idx ON category(verify bool_ops);

CREATE TABLE comment (
    id character varying DEFAULT uuid_generate_v1() PRIMARY KEY,
    post_id character varying NOT NULL,
    owner_id character varying(40) NOT NULL,
    owner_type integer NOT NULL,
    text text NOT NULL,
    created_at integer NOT NULL,
    updated_at integer NOT NULL,
    status boolean NOT NULL,
    used boolean NOT NULL
);

CREATE UNIQUE INDEX comment_pkey ON comment(id text_ops);
CREATE INDEX comment_id_idx ON comment(id text_ops);
CREATE INDEX comment_post_id_idx ON comment(post_id text_ops);
CREATE INDEX comment_owner_id_idx ON comment(owner_id text_ops);
CREATE INDEX comment_created_at_idx ON comment(created_at int4_ops);
CREATE INDEX comment_status_idx ON comment(status bool_ops);
CREATE INDEX comment_used_idx ON comment(used bool_ops);

CREATE TABLE contact (
    id character varying DEFAULT uuid_generate_v1() PRIMARY KEY,
    owner_id character varying NOT NULL,
    owner_type integer NOT NULL,
    tel character varying,
    website character varying,
    social character varying,
    email character varying,
    address text,
    city character varying,
    country character varying,
    created_at character varying,
    updated_at character varying
);

CREATE UNIQUE INDEX contact_pkey ON contact(id text_ops);
CREATE INDEX contact_owner_id_idx ON contact(owner_id text_ops);
CREATE INDEX contact_owner_type_idx ON contact(owner_type int4_ops);

CREATE TABLE follow_gap (
    id character varying DEFAULT uuid_generate_v1() PRIMARY KEY,
    created_at character varying NOT NULL,
    owner_id character varying NOT NULL,
    gap_id character varying NOT NULL,
    status boolean NOT NULL,
    used boolean NOT NULL,
    updated_at character varying NOT NULL
);

CREATE UNIQUE INDEX follow_gap_pkey ON follow_gap(id text_ops);
CREATE INDEX follow_gap_created_at_idx ON follow_gap(created_at text_ops);
CREATE INDEX follow_gap_owner_id_idx ON follow_gap(owner_id text_ops);
CREATE INDEX follow_gap_gap_id_idx ON follow_gap(gap_id text_ops);
CREATE INDEX follow_gap_status_idx ON follow_gap(status bool_ops);
CREATE INDEX follow_gap_used_idx ON follow_gap(used bool_ops);
CREATE INDEX follow_gap_id_idx ON follow_gap(id text_ops);

CREATE TABLE follow_topic (
    id character varying DEFAULT uuid_generate_v1() PRIMARY KEY,
    created_at character varying NOT NULL,
    owner_id character varying NOT NULL,
    status boolean NOT NULL,
    topic_id character varying NOT NULL,
    used boolean NOT NULL,
    updated_at character varying
);

CREATE UNIQUE INDEX follow_topic_pkey ON follow_topic(id text_ops);
CREATE INDEX follow_topic_id_idx ON follow_topic(id text_ops);
CREATE INDEX follow_topic_created_at_idx ON follow_topic(created_at text_ops);
CREATE INDEX follow_topic_owner_id_idx ON follow_topic(owner_id text_ops);
CREATE INDEX follow_topic_status_idx ON follow_topic(status bool_ops);
CREATE INDEX follow_topic_topic_id_idx ON follow_topic(topic_id text_ops);
CREATE INDEX follow_topic_used_idx ON follow_topic(used bool_ops);

CREATE TABLE gap (
    id character varying DEFAULT uuid_generate_v1() PRIMARY KEY,
    bio text,
    cat_id character varying NOT NULL,
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

CREATE UNIQUE INDEX gap_pkey ON gap(id text_ops);
CREATE INDEX gap_id_idx ON gap(id text_ops);
CREATE INDEX gap_used_idx ON gap(used bool_ops);
CREATE INDEX gap_user_id_idx ON gap(user_id text_ops);

CREATE TABLE gap_audit (
    id character varying DEFAULT uuid_generate_v1() PRIMARY KEY,
    created_at character varying NOT NULL,
    gap_id character varying NOT NULL,
    snapshot json NOT NULL,
    type character(10) NOT NULL,
    updated_at character varying NOT NULL
);

CREATE UNIQUE INDEX gap_audit_pkey ON gap_audit(id text_ops);

CREATE TABLE gap_recommend (
    id character varying DEFAULT uuid_generate_v1() PRIMARY KEY,
    gap_id character varying,
    created_at character varying
);

CREATE UNIQUE INDEX gap_recommend_pkey ON gap_recommend(id text_ops);
CREATE INDEX gap_recommend_gap_id_idx ON gap_recommend(gap_id text_ops);
CREATE INDEX gap_recommend_created_at_idx ON gap_recommend(created_at text_ops);

CREATE TABLE gap_view (
    id character varying DEFAULT uuid_generate_v1() PRIMARY KEY,
    created_at character varying,
    owner_id character varying,
    gap_id character varying,
    updated_at character varying
);

CREATE UNIQUE INDEX gap_view_pkey ON gap_view(id text_ops);

CREATE TABLE general_stat (
    id character varying DEFAULT uuid_generate_v1() PRIMARY KEY,
    key character varying NOT NULL,
    value character varying NOT NULL,
    created_at character varying NOT NULL
);

CREATE UNIQUE INDEX general_stat_pkey ON general_stat(id text_ops);

CREATE TABLE guest (
    id character varying DEFAULT uuid_generate_v1() PRIMARY KEY,
    post_id character varying NOT NULL,
    visitor_id character varying NOT NULL,
    user_agent character varying NOT NULL,
    referrer character varying NOT NULL,
    created_at integer NOT NULL,
    updated_at integer NOT NULL
);

CREATE UNIQUE INDEX guest_pkey ON guest(id text_ops);
CREATE INDEX guest_created_at_idx ON guest(created_at int4_ops);
CREATE INDEX guest_post_id_idx ON guest(post_id text_ops);
CREATE INDEX guest_visitor_id_idx ON guest(visitor_id text_ops);

CREATE TABLE imageinbucket (
    id character varying DEFAULT uuid_generate_v1() PRIMARY KEY,
    owner_id character varying(40) NOT NULL,
    image text NOT NULL,
    height integer NOT NULL DEFAULT 0,
    width integer NOT NULL DEFAULT 0,
    status boolean NOT NULL,
    created_at integer NOT NULL
);

CREATE UNIQUE INDEX imageinbucket_pkey ON imageinbucket(id text_ops);
CREATE INDEX imageinbucket_owner_id_idx ON imageinbucket(owner_id text_ops);
CREATE INDEX imageinbucket_image_idx ON imageinbucket(image text_ops);
CREATE INDEX imageinbucket_status_idx ON imageinbucket(status bool_ops);

CREATE TABLE likepost (
    id character varying DEFAULT uuid_generate_v1() PRIMARY KEY,
    post_id character varying NOT NULL,
    owner_id character varying(40) NOT NULL,
    created_at integer NOT NULL,
    updated_at integer NOT NULL,
    status boolean NOT NULL,
    used boolean NOT NULL
);

CREATE UNIQUE INDEX likepost_pkey ON likepost(id text_ops);
CREATE INDEX likepost_post_id_idx ON likepost(post_id text_ops);
CREATE INDEX likepost_owner_id_idx ON likepost(owner_id text_ops);
CREATE INDEX likepost_status_idx ON likepost(status bool_ops);
CREATE INDEX likepost_used_idx ON likepost(used bool_ops);

CREATE TABLE notification (
    id character varying DEFAULT uuid_generate_v1() PRIMARY KEY,
    owner_id character varying NOT NULL,
    user_id character varying NOT NULL,
    post_id character varying NOT NULL,
    gap_id character varying NOT NULL,
    type integer NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT now(),
    updated_at timestamp without time zone NOT NULL DEFAULT now(),
    read boolean NOT NULL,
    main boolean NOT NULL,
    count integer NOT NULL,
    used boolean NOT NULL,
    comment_text text NOT NULL DEFAULT ''::text
);

CREATE UNIQUE INDEX untitled_table_pkey1 ON notification(id text_ops);
CREATE INDEX notification_id_idx ON notification(id text_ops);
CREATE INDEX notification_owner_id_idx ON notification(owner_id text_ops);
CREATE INDEX notification_user_id_idx ON notification(user_id text_ops);
CREATE INDEX notification_post_id_idx ON notification(post_id text_ops);
CREATE INDEX notification_gap_id_idx ON notification(gap_id text_ops);
CREATE INDEX notification_type_idx ON notification(type int4_ops);
CREATE INDEX notification_created_at_idx ON notification(created_at timestamp_ops);
CREATE INDEX notification_read_idx ON notification(read bool_ops);
CREATE INDEX notification_main_idx ON notification(main bool_ops);
CREATE INDEX notification_count_idx ON notification(count int4_ops);
CREATE INDEX notification_used_idx ON notification(used bool_ops);

CREATE TABLE post (
    id character varying DEFAULT uuid_generate_v1() PRIMARY KEY,
    slug character varying(60) NOT NULL,
    owner_id character varying(40) NOT NULL,
    owner_type integer NOT NULL,
    title text NOT NULL,
    description text NOT NULL,
    image_url text NOT NULL,
    height integer NOT NULL DEFAULT 0,
    width integer NOT NULL DEFAULT 0,
    link text NOT NULL,
    link_description text NOT NULL,
    created_at integer NOT NULL,
    updated_at integer NOT NULL,
    type integer NOT NULL,
    province character varying(10) NOT NULL,
    like_count integer NOT NULL DEFAULT 0,
    comment_count integer NOT NULL DEFAULT 0,
    share_count integer NOT NULL DEFAULT 0,
    view_count integer NOT NULL DEFAULT 0,
    status boolean NOT NULL,
    used boolean NOT NULL,
    user_id character varying,
    draft boolean NOT NULL DEFAULT false,
    verify integer NOT NULL DEFAULT 0,
    image_share_url text,
    width_share integer DEFAULT 0,
    height_share integer DEFAULT 0,
    guest_view_count integer NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX post_pkey ON post(id text_ops);
CREATE INDEX post_id_idx ON post(id text_ops);
CREATE INDEX post_slug_idx ON post(slug text_ops);
CREATE INDEX post_owner_id_idx ON post(owner_id text_ops);
CREATE INDEX post_created_at_idx ON post(created_at int4_ops);
CREATE INDEX post_type_idx ON post(type int4_ops);
CREATE INDEX post_status_idx ON post(status bool_ops);
CREATE INDEX post_draft_idx ON post(draft bool_ops);
CREATE INDEX post_used_idx ON post(used bool_ops);
CREATE INDEX post_user_id_idx ON post(user_id text_ops);
CREATE INDEX post_verify_idx ON post(verify int4_ops);

CREATE TABLE statistic_token (
    id character varying DEFAULT uuid_generate_v1() PRIMARY KEY,
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

CREATE UNIQUE INDEX statistic_token_pkey ON statistic_token(id text_ops);

CREATE TABLE tagtopic (
    id character varying DEFAULT uuid_generate_v1() PRIMARY KEY,
    post_id character varying NOT NULL,
    topic_id character varying(40) NOT NULL,
    created_at integer NOT NULL,
    main boolean NOT NULL DEFAULT false
);

CREATE UNIQUE INDEX tagtopic_pkey ON tagtopic(id text_ops);
CREATE INDEX tagtopic_id_idx ON tagtopic(id text_ops);
CREATE INDEX tagtopic_post_id_idx ON tagtopic(post_id text_ops);
CREATE INDEX tagtopic_topic_id_idx ON tagtopic(topic_id text_ops);
CREATE INDEX tagtopic_created_at_idx ON tagtopic(created_at int4_ops);
CREATE INDEX tagtopic_main_idx ON tagtopic(main bool_ops);

CREATE TABLE topic (
    id character varying DEFAULT uuid_generate_v1() PRIMARY KEY,
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

CREATE UNIQUE INDEX topic_pkey ON topic(id text_ops);
CREATE INDEX topic_code_idx ON topic(code text_ops);
CREATE INDEX topic_cat_id_idx ON topic(cat_id text_ops);
CREATE INDEX topic_verify_idx ON topic(verify bool_ops);
CREATE INDEX topic_used_count_idx ON topic(used_count int4_ops);
CREATE INDEX topic_count_idx ON topic(count int4_ops);

CREATE TABLE user (
    id character varying DEFAULT uuid_generate_v1() PRIMARY KEY,
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
    verify json NOT NULL,
    notification boolean NOT NULL DEFAULT false
);

CREATE UNIQUE INDEX user_pkey ON user(id text_ops);
CREATE INDEX user_id_idx ON user(id text_ops);
CREATE INDEX user_used_idx ON user(used bool_ops);

CREATE TABLE user_audit (
    id character varying DEFAULT uuid_generate_v1() PRIMARY KEY,
    created_at character varying NOT NULL,
    user_id character varying NOT NULL,
    snapshot json NOT NULL,
    type character(10) NOT NULL,
    updated_at character varying NOT NULL
);

CREATE UNIQUE INDEX untitled_table_pkey ON user_audit(id text_ops);

CREATE TABLE user_provider (
    id character varying DEFAULT uuid_generate_v1() PRIMARY KEY,
    code_social_signin text,
    count integer,
    created_at character varying,
    provider json,
    time character varying,
    update_email json,
    updated_at character varying
);

CREATE UNIQUE INDEX user_provider_pkey ON user_provider(id text_ops);
CREATE INDEX user_provider_count_idx ON user_provider(count int4_ops);
CREATE INDEX user_provider_time_idx ON user_provider(time text_ops);

CREATE TABLE view (
    id character varying NOT NULL DEFAULT uuid_generate_v1(),
    post_id character varying NOT NULL,
    owner_id character varying(40) NOT NULL,
    created_at character varying NOT NULL
);

