-- # Table: users
-- DROP TABLE public.users;
CREATE TABLE public.users (
	id uuid NOT NULL,
	username varchar(50) NOT NULL,
	"name" varchar(255) NOT NULL,
	"password" varchar(255) NOT NULL,
	created_at timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at timestamp(0) NULL,
	CONSTRAINT users_pkey PRIMARY KEY (id),
	CONSTRAINT users_username_unique UNIQUE (username)
);
CREATE INDEX users_username_index ON public.users USING btree (username);

-- # Table: logs
-- DROP TABLE public.logs;
CREATE TABLE public.user_logs (
	id uuid NOT NULL,
	message text NOT NULL,
	channel varchar(255) NOT NULL,
	"level" int2 NOT NULL DEFAULT '0'::smallint,
	level_name varchar(20) NOT NULL,
	datetime varchar(255) NOT NULL,
	context text NOT NULL,
	extra text NOT NULL,
	created_at timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at timestamp(0) NULL
);

-- # Table: items
-- DROP TABLE public.items;
CREATE TABLE public.items (
	id uuid NOT NULL,
	code varchar(10) NOT NULL,
  "name" varchar(255) NOT NULL,
  qty float NOT NULL,
  price float NOT NULL,
	created_at timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at timestamp(0) NULL,
  deleted_at timestamp(0) NULL,
	created_by uuid NULL,
	updated_by uuid NULL,
	CONSTRAINT items_code_unique UNIQUE (code),
	CONSTRAINT items_pkey PRIMARY KEY (id),
	CONSTRAINT items_created_by_foreign FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE SET NULL,
	CONSTRAINT items_updated_by_foreign FOREIGN KEY (updated_by) REFERENCES public.users(id) ON DELETE SET NULL
);

-- # Table: stock_mutations
-- DROP TABLE public.stock_mutations;
CREATE TABLE public.stock_mutations (
	id uuid NOT NULL,
  item_id uuid NOT NULL,
  period varchar(50) NOT NULL,
  trx_date date NOT NULL,
  qty float NOT NULL,
  model_id uuid NOT NULL,
  model_type varchar(255) NOT NULL,
  created_at timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at timestamp(0) NULL,
	CONSTRAINT stock_mutations_pkey PRIMARY KEY (id),
  CONSTRAINT stock_mutations_item_id_foreign FOREIGN KEY (item_id) REFERENCES public.items(id) ON DELETE CASCADE
);

-- # Table: purchase_headers
-- DROP TABLE public.purchase_headers;
CREATE TABLE public.purchase_headers (
	id uuid NOT NULL,
  purchase_invoice_num varchar(255) NOT NULL,
  purchase_date date NOT NULL,
  supplier_name varchar(255) NOT NULL,
  created_at timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at timestamp(0) NULL,
  created_by uuid NULL,
  updated_by uuid NULL,
  CONSTRAINT purchase_headers_pkey PRIMARY KEY (id),
  CONSTRAINT purchase_headers_created_by_foreign FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE SET NULL,
  CONSTRAINT purchase_headers_updated_by_foreign FOREIGN KEY (updated_by) REFERENCES public.users(id) ON DELETE SET NULL
);

-- # Table: purchase_details
-- DROP TABLE public.purchase_details;
CREATE TABLE public.purchase_details (
	id uuid NOT NULL,
  header_id uuid NOT NULL,
  item_id uuid NOT NULL,
  qty float NOT NULL,
  price_amount float NOT NULL,
  total_amount float NOT NULL,
  created_at timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp(0) NULL,
  CONSTRAINT purchase_details_pkey PRIMARY KEY (id),
  CONSTRAINT purchase_details_header_id_foreign FOREIGN KEY (header_id) REFERENCES public.purchase_headers(id) ON DELETE CASCADE,
  CONSTRAINT purchase_details_item_id_foreign FOREIGN KEY (item_id) REFERENCES public.items(id) ON DELETE CASCADE
);

-- # Table: sell_headers
-- DROP TABLE public.sell_headers;
CREATE TABLE public.sell_headers (
	id uuid NOT NULL,
  sell_invoice_num varchar(255) NOT NULL,
  sell_date date NOT NULL,
  customer_name varchar(255) NOT NULL,
  created_at timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at timestamp(0) NULL,
  created_by uuid NULL,
  updated_by uuid NULL,
  CONSTRAINT sell_headers_pkey PRIMARY KEY (id),
  CONSTRAINT sell_headers_created_by_foreign FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE SET NULL,
  CONSTRAINT sell_headers_updated_by_foreign FOREIGN KEY (updated_by) REFERENCES public.users(id) ON DELETE SET NULL
);

-- # Table: sell_details
-- DROP TABLE public.sell_details;
CREATE TABLE public.sell_details (
	id uuid NOT NULL,
  header_id uuid NOT NULL,
  item_id uuid NOT NULL,
  qty float NOT NULL,
  price_amount float NOT NULL,
  total_amount float NOT NULL,
  created_at timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp(0) NULL,
  CONSTRAINT sell_details_pkey PRIMARY KEY (id),
  CONSTRAINT sell_details_header_id_foreign FOREIGN KEY (header_id) REFERENCES public.sell_headers(id) ON DELETE CASCADE,
  CONSTRAINT sell_details_item_id_foreign FOREIGN KEY (item_id) REFERENCES public.items(id) ON DELETE CASCADE
);
