-- Etapa 1 — Fundação
-- Extensões, função utilitária de updated_at e categorias profissionais.

CREATE EXTENSION IF NOT EXISTS pgcrypto;   -- gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS citext;     -- e-mails case-insensitive (Etapa 2)
CREATE EXTENSION IF NOT EXISTS btree_gist; -- exclusion constraint de conflito de horários (Etapa 5)

-- Mantém updated_at correto mesmo em UPDATEs feitos fora da aplicação.
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE professional_categories (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    name        text        NOT NULL CHECK (char_length(btrim(name)) BETWEEN 2 AND 80),
    slug        text        NOT NULL CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    description text        CHECK (description IS NULL OR char_length(description) <= 500),
    is_active   boolean     NOT NULL DEFAULT true,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

-- Unicidade apenas entre registros não excluídos (exclusão lógica).
CREATE UNIQUE INDEX ux_professional_categories_slug
    ON professional_categories (slug) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX ux_professional_categories_name
    ON professional_categories (lower(name)) WHERE deleted_at IS NULL;

CREATE TRIGGER trg_professional_categories_updated_at
    BEFORE UPDATE ON professional_categories
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

INSERT INTO professional_categories (name, slug, description) VALUES
    ('Garçom / Garçonete',               'garcom',                'Atendimento e serviço de mesas.'),
    ('Bartender',                        'bartender',             'Preparo de drinks autorais e coquetelaria.'),
    ('Barman',                           'barman',                'Operação de bar e serviço de bebidas.'),
    ('Barback',                          'barback',               'Apoio ao bar: reposição, gelo e limpeza.'),
    ('Cozinheiro',                       'cozinheiro',            'Preparo de pratos e produção de cozinha.'),
    ('Auxiliar de cozinha',              'auxiliar-de-cozinha',   'Pré-preparo e apoio à cozinha.'),
    ('Recepcionista',                    'recepcionista',         'Recepção e credenciamento de clientes.'),
    ('Hostess',                          'hostess',               'Acolhimento e condução de convidados.'),
    ('Segurança',                        'seguranca',             'Controle de acesso e segurança patrimonial.'),
    ('Caixa',                            'caixa',                 'Operação de caixa e fichas.'),
    ('Auxiliar de limpeza',              'auxiliar-de-limpeza',   'Limpeza e conservação do espaço.'),
    ('Promotor de eventos',              'promotor-de-eventos',   'Divulgação e ativação de público.'),
    ('Coordenador de eventos',           'coordenador-de-eventos','Coordenação de equipes e operação do evento.'),
    ('Montador de estrutura',            'montador-de-estrutura', 'Montagem e desmontagem de estruturas.'),
    ('Staff de eventos',                 'staff-de-eventos',      'Apoio operacional geral.'),
    ('DJ e profissionais de entretenimento', 'dj-entretenimento', 'Música e entretenimento.');
