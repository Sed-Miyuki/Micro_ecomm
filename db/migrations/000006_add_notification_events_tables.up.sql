CREATE TYPE notification_state_enum AS ENUM('not sent','sent','failed');

CREATE TABLE notification_states(
    id              BIGSERIAL                   PRIMARY KEY,
    order_id        INT                         NOT NULL,
    state           notification_state_enum     NOT NULL,
    message         VARCHAR(512),
    requested_at    timestamptz                 DEFAULT now(),
    completed_at    timestamptz
);

CREATE TABLE notification_events_queue(
    id              BIGSERIAL                   PRIMARY KEY,
    user_email      VARCHAR(256)                NOT NULL,
    order_status    VARCHAR(256)                NOT NULL,
    order_id        INT                         NOT NULL,
    state_id        INT                         NOT NULL,
    attempts        INT,
    created_at      timestamptz                 DEFAULT NOW(),
    updated_at      timestamptz
);

ALTER TABLE notification_states
    ADD CONSTRAINT notification_states_order_id_fk 
    FOREIGN KEY (order_id) REFERENCES orders (id);

ALTER TABLE notification_events_queue
    ADD CONSTRAINT notification_events_queue_order_id_fk 
    FOREIGN KEY (order_id) REFERENCES orders (id),
    ADD CONSTRAINT notification_events_queue_state_id_fk 
    FOREIGN KEY (state_id) REFERENCES notification_states (id);