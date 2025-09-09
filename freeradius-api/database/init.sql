-- FreeRADIUS Database Initialization Script
-- This script creates the necessary tables for FreeRADIUS API

-- Create database (uncomment if needed)
-- CREATE DATABASE IF NOT EXISTS radius;
-- USE radius;

-- Drop tables if they exist (for clean installation)
DROP TABLE IF EXISTS radusergroup;
DROP TABLE IF EXISTS radgroupreply;
DROP TABLE IF EXISTS radgroupcheck;
DROP TABLE IF EXISTS radreply;
DROP TABLE IF EXISTS radcheck;
DROP TABLE IF EXISTS nas;

-- Create nas table
CREATE TABLE nas (
  id int(10) NOT NULL AUTO_INCREMENT,
  nasname varchar(128) NOT NULL,
  shortname varchar(32),
  type varchar(30) DEFAULT 'other',
  ports int(5),
  secret varchar(60) DEFAULT 'secret' NOT NULL,
  server varchar(64),
  community varchar(50),
  description varchar(200) DEFAULT 'RADIUS Client',
  PRIMARY KEY (id),
  KEY nasname (nasname)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create radcheck table
CREATE TABLE radcheck (
  id int(11) unsigned NOT NULL AUTO_INCREMENT,
  username varchar(64) NOT NULL DEFAULT '',
  attribute varchar(64) NOT NULL DEFAULT '',
  op char(2) NOT NULL DEFAULT '==',
  value varchar(253) NOT NULL DEFAULT '',
  PRIMARY KEY (id),
  KEY username (username(32))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create radreply table
CREATE TABLE radreply (
  id int(11) unsigned NOT NULL AUTO_INCREMENT,
  username varchar(64) NOT NULL DEFAULT '',
  attribute varchar(64) NOT NULL DEFAULT '',
  op char(2) NOT NULL DEFAULT '=',
  value varchar(253) NOT NULL DEFAULT '',
  PRIMARY KEY (id),
  KEY username (username(32))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create radgroupcheck table (optional, for group-based authentication)
CREATE TABLE radgroupcheck (
  id int(11) unsigned NOT NULL AUTO_INCREMENT,
  groupname varchar(64) NOT NULL DEFAULT '',
  attribute varchar(64) NOT NULL DEFAULT '',
  op char(2) NOT NULL DEFAULT '==',
  value varchar(253) NOT NULL DEFAULT '',
  PRIMARY KEY (id),
  KEY groupname (groupname(32))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create radgroupreply table (optional, for group-based authentication)
CREATE TABLE radgroupreply (
  id int(11) unsigned NOT NULL AUTO_INCREMENT,
  groupname varchar(64) NOT NULL DEFAULT '',
  attribute varchar(64) NOT NULL DEFAULT '',
  op char(2) NOT NULL DEFAULT '=',
  value varchar(253) NOT NULL DEFAULT '',
  PRIMARY KEY (id),
  KEY groupname (groupname(32))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create radusergroup table (optional, for user-group mapping)
CREATE TABLE radusergroup (
  username varchar(64) NOT NULL DEFAULT '',
  groupname varchar(64) NOT NULL DEFAULT '',
  priority int(11) NOT NULL DEFAULT '1',
  KEY username (username(32))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert sample NAS entries
INSERT INTO nas (nasname, shortname, type, ports, secret, description) VALUES
('127.0.0.1', 'localhost', 'other', 1812, 'testing123', 'Local test server'),
('192.168.1.1', 'router1', 'cisco', 1812, 'secret123', 'Main router'),
('192.168.1.10', 'switch1', 'cisco', 1812, 'switch_secret', 'Main switch'),
('10.0.0.1', 'firewall1', 'other', 1812, 'fw_secret', 'Main firewall');

-- Insert sample user entries
-- Admin user
INSERT INTO radcheck (username, attribute, op, value) VALUES
('admin', 'Cleartext-Password', ':=', 'admin123!'),
('testuser1', 'Cleartext-Password', ':=', 'testpass123'),
('testuser2', 'Cleartext-Password', ':=', 'userpass456'),
('pppuser1', 'Cleartext-Password', ':=', 'ppp123'),
('dialup1', 'Cleartext-Password', ':=', 'dialup456');

-- User reply attributes (profiles)
INSERT INTO radreply (username, attribute, op, value) VALUES
('admin', 'Framed-Protocol', ':=', 'PPP'),
('admin', 'Service-Type', ':=', 'Framed-User'),
('testuser1', 'Framed-Protocol', ':=', 'PPP'),
('testuser1', 'Service-Type', ':=', 'Framed-User'),
('testuser2', 'Framed-Protocol', ':=', 'SLIP'),
('testuser2', 'Service-Type', ':=', 'Framed-User'),
('pppuser1', 'Framed-Protocol', ':=', 'PPP'),
('pppuser1', 'Service-Type', ':=', 'Framed-User'),
('pppuser1', 'Framed-IP-Address', ':=', '192.168.100.10'),
('dialup1', 'Framed-Protocol', ':=', 'PPP'),
('dialup1', 'Service-Type', ':=', 'Framed-User'),
('dialup1', 'Session-Timeout', ':=', '3600');

-- Insert sample group entries (optional)
INSERT INTO radgroupcheck (groupname, attribute, op, value) VALUES
('ppp_users', 'Auth-Type', ':=', 'Local'),
('dialup_users', 'Auth-Type', ':=', 'Local'),
('admin_users', 'Auth-Type', ':=', 'Local');

INSERT INTO radgroupreply (groupname, attribute, op, value) VALUES
('ppp_users', 'Framed-Protocol', ':=', 'PPP'),
('ppp_users', 'Service-Type', ':=', 'Framed-User'),
('dialup_users', 'Framed-Protocol', ':=', 'PPP'),
('dialup_users', 'Service-Type', ':=', 'Framed-User'),
('dialup_users', 'Session-Timeout', ':=', '7200'),
('admin_users', 'Framed-Protocol', ':=', 'PPP'),
('admin_users', 'Service-Type', ':=', 'Administrative-User');

-- Map users to groups (optional)
INSERT INTO radusergroup (username, groupname, priority) VALUES
('testuser1', 'ppp_users', 1),
('testuser2', 'ppp_users', 1),
('pppuser1', 'ppp_users', 1),
('dialup1', 'dialup_users', 1),
('admin', 'admin_users', 1);

-- Create indexes for better performance
CREATE INDEX idx_radcheck_username ON radcheck(username);
CREATE INDEX idx_radcheck_username_attribute ON radcheck(username, attribute);
CREATE INDEX idx_radreply_username ON radreply(username);
CREATE INDEX idx_radreply_username_attribute ON radreply(username, attribute);
CREATE INDEX idx_nas_nasname ON nas(nasname);
CREATE INDEX idx_nas_shortname ON nas(shortname);
CREATE INDEX idx_radgroupcheck_groupname ON radgroupcheck(groupname);
CREATE INDEX idx_radgroupreply_groupname ON radgroupreply(groupname);
CREATE INDEX idx_radusergroup_username ON radusergroup(username);
CREATE INDEX idx_radusergroup_groupname ON radusergroup(groupname);

-- Show table status
SELECT 'Database initialization completed successfully!' as status;
SELECT COUNT(*) as nas_count FROM nas;
SELECT COUNT(*) as users_count FROM radcheck;
SELECT COUNT(*) as reply_attributes_count FROM radreply;

-- Display sample data
SELECT 'Sample NAS entries:' as info;
SELECT id, nasname, shortname, type, description FROM nas LIMIT 5;

SELECT 'Sample user entries:' as info;
SELECT username, attribute, value FROM radcheck WHERE attribute = 'Cleartext-Password' LIMIT 5;

SELECT 'Sample user profiles:' as info;
SELECT username, attribute, value FROM radreply LIMIT 10;