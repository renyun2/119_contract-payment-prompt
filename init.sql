-- 合同台账与付款节点：表结构 + 演示数据（50 份合同、多阶段节点、历史付款与催收跟进）

DROP TABLE IF EXISTS collection_followups CASCADE;
DROP TABLE IF EXISTS actual_payments CASCADE;
DROP TABLE IF EXISTS payment_nodes CASCADE;
DROP TABLE IF EXISTS contracts CASCADE;
DROP TYPE IF EXISTS contract_status CASCADE;
DROP TYPE IF EXISTS contract_type CASCADE;

CREATE TYPE contract_type AS ENUM ('purchase', 'sales', 'service', 'engineering');
CREATE TYPE contract_status AS ENUM ('active', 'completed', 'terminated');

CREATE TABLE contracts (
  id SERIAL PRIMARY KEY,
  contract_no VARCHAR(64) NOT NULL UNIQUE,
  title VARCHAR(256) NOT NULL,
  signed_date DATE NOT NULL,
  type contract_type NOT NULL,
  counterparty VARCHAR(256) NOT NULL,
  total_amount NUMERIC(18, 2) NOT NULL,
  period_start DATE,
  period_end DATE,
  summary TEXT,
  status contract_status NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE payment_nodes (
  id SERIAL PRIMARY KEY,
  contract_id INT NOT NULL REFERENCES contracts (id) ON DELETE CASCADE,
  node_name VARCHAR(128) NOT NULL,
  trigger_condition TEXT,
  amount NUMERIC(18, 2) NOT NULL,
  planned_date DATE NOT NULL,
  is_triggered BOOLEAN NOT NULL DEFAULT FALSE,
  is_paid BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE actual_payments (
  id SERIAL PRIMARY KEY,
  contract_id INT NOT NULL REFERENCES contracts (id) ON DELETE CASCADE,
  node_id INT NOT NULL REFERENCES payment_nodes (id) ON DELETE CASCADE,
  pay_date DATE NOT NULL,
  amount NUMERIC(18, 2) NOT NULL,
  bank_ref VARCHAR(128),
  pay_account VARCHAR(128),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE collection_followups (
  id SERIAL PRIMARY KEY,
  contract_id INT NOT NULL REFERENCES contracts (id) ON DELETE CASCADE,
  node_id INT REFERENCES payment_nodes (id) ON DELETE SET NULL,
  follower VARCHAR(64) NOT NULL,
  follow_date DATE NOT NULL,
  content TEXT NOT NULL,
  promised_pay_date DATE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payment_nodes_contract ON payment_nodes (contract_id);
CREATE INDEX idx_payment_nodes_planned ON payment_nodes (planned_date);
CREATE INDEX idx_payment_nodes_unpaid ON payment_nodes (is_paid, planned_date);
CREATE INDEX idx_actual_payments_contract ON actual_payments (contract_id);
CREATE INDEX idx_collection_followups_contract ON collection_followups (contract_id);

-- 50 份合同
INSERT INTO contracts (contract_no, title, signed_date, type, counterparty, total_amount, period_start, period_end, summary, status)
SELECT
  'CT-2026-' || LPAD(g::TEXT, 3, '0'),
  CASE ((g + 2) % 4)
    WHEN 0 THEN '设备与物资采购协议-' || g
    WHEN 1 THEN '产品销售框架合同-' || g
    WHEN 2 THEN '运维与技术支持服务-' || g
    ELSE '工程施工总承包补充-' || g
  END,
  (DATE '2025-02-01' + ((g * 17) % 420))::date,
  (ARRAY['purchase','sales','service','engineering']::contract_type[])[1 + ((g + 3) % 4)],
  '客商主体-' || LPAD(((g * 73) % 500 + 1)::TEXT, 3, '0'),
  ROUND((88000 + (g * 17431) % 4200000)::numeric / 100, 2),
  (DATE '2025-02-01' + ((g * 17) % 420))::date,
  (DATE '2025-02-01' + ((g * 17) % 420) + 400 + ((g % 9) * 30))::date,
  '演示摘要：标的物/服务范围见附件清单，含税总价以本表为准。流水号 #' || g,
  CASE
    WHEN g % 8 = 0 THEN 'completed'::contract_status
    WHEN g % 13 = 0 THEN 'terminated'::contract_status
    ELSE 'active'::contract_status
  END
FROM generate_series(1, 50) AS g;

-- 每份合同 4 个付款/收款节点；部分节点计划在 2026-05 月内到期便于首页演示
DO $$
DECLARE
  c RECORD;
  parts text[] := ARRAY['预付款', '进度款', '尾款', '质保金'];
  pcts numeric[] := ARRAY[0.20, 0.45, 0.30, 0.05];
  days_base int[] := ARRAY[10, 95, 185, 380];
  i int;
  p date;
  amt numeric;
BEGIN
  FOR c IN SELECT * FROM contracts ORDER BY id LOOP
    FOR i IN 1..4 LOOP
      p := c.signed_date + days_base[i];
      IF c.id BETWEEN 10 AND 40 AND i IN (2, 3) THEN
        p := DATE '2026-05-01' + ((c.id * 7 + i * 11) % 31);
      ELSIF c.id BETWEEN 41 AND 50 AND i = 1 THEN
        p := DATE '2026-05-06' + (c.id % 12);
      END IF;

      amt := ROUND(c.total_amount * pcts[i], 2);
      INSERT INTO payment_nodes (contract_id, node_name, trigger_condition, amount, planned_date, is_triggered, is_paid)
      VALUES (
        c.id,
        parts[i],
        '合同约定：到货/里程碑/验收/质保条件满足后触发',
        amt,
        p,
        p < DATE '2026-03-01',
        FALSE
      );
    END LOOP;
  END LOOP;
END $$;

-- 历史实际付款（部分节点已付）
UPDATE payment_nodes n
SET is_paid = TRUE, is_triggered = TRUE
FROM contracts c
WHERE n.contract_id = c.id
  AND (
    (n.node_name = '预付款' AND c.id % 2 = 0)
    OR (n.node_name = '进度款' AND c.id % 5 IN (1, 2))
    OR (n.node_name = '尾款' AND c.id % 11 = 0)
  );

INSERT INTO actual_payments (contract_id, node_id, pay_date, amount, bank_ref, pay_account)
SELECT
  n.contract_id,
  n.id,
  (n.planned_date - ((n.id % 9) + 3))::date,
  n.amount,
  'BK' || LPAD(n.id::text, 10, '0'),
  CASE (n.contract_id % 3)
    WHEN 0 THEN '工行**科技园支行'
    WHEN 1 THEN '建行**分行营业部'
    ELSE '招行**监管专户'
  END
FROM payment_nodes n
WHERE n.is_paid = TRUE;

-- 销售合同逾期未收节点：催收跟进记录
INSERT INTO collection_followups (contract_id, node_id, follower, follow_date, content, promised_pay_date)
SELECT
  n.contract_id,
  n.id,
  CASE (n.id % 4) WHEN 0 THEN '李芳' WHEN 1 THEN '王磊' WHEN 2 THEN '赵敏' ELSE '周洋' END,
  DATE '2026-04-05' + ((n.id % 18)),
  CASE (n.id % 3)
    WHEN 0 THEN '与对方财务确认开票及付款批次，已进入审批。'
    WHEN 1 THEN '法务已发催收函副本，业务部门持续沟通。'
    ELSE '客户经理上门沟通，争取本周内首付款到账。'
  END,
  DATE '2026-05-28' + ((n.id % 12))
FROM payment_nodes n
JOIN contracts c ON c.id = n.contract_id
WHERE c.type = 'sales'
  AND n.is_paid = FALSE
  AND n.planned_date < DATE '2026-05-10'
  AND (n.contract_id % 4) = (n.id % 4);
