const db = require('../config/database');

class NasModel {
  // Get all NAS entries
  static async getAll() {
    const query = 'SELECT id, shortname, nasname, secret, type, ports, community, description FROM nas ORDER BY shortname';
    return await db.query(query);
  }

  // Get NAS by ID
  static async getById(id) {
    const query = 'SELECT id, shortname, nasname, secret, type, ports, community, description FROM nas WHERE id = ?';
    const result = await db.query(query, [id]);
    return result[0] || null;
  }

  // Get NAS by shortname
  static async getByShortname(shortname) {
    const query = 'SELECT id, shortname, nasname, secret, type, ports, community, description FROM nas WHERE shortname = ?';
    const result = await db.query(query, [shortname]);
    return result[0] || null;
  }

  // Get NAS by IP address (nasname)
  static async getByIp(nasname) {
    const query = 'SELECT id, shortname, nasname, secret, type, ports, community, description FROM nas WHERE nasname = ?';
    const result = await db.query(query, [nasname]);
    return result[0] || null;
  }

  // Create new NAS
  static async create(nasData) {
    const { name, ip, secret, type = 'other', ports = 1812, community = '', description = '' } = nasData;
    
    const query = `
      INSERT INTO nas (shortname, nasname, secret, type, ports, community, description) 
      VALUES (?, ?, ?, ?, ?, ?, ?)
    `;
    
    const result = await db.query(query, [name, ip, secret, type, ports, community, description]);
    return {
      id: result.insertId,
      shortname: name,
      nasname: ip,
      secret,
      type,
      ports,
      community,
      description
    };
  }

  // Update NAS
  static async update(id, nasData) {
    const { name, ip, secret, type, ports, community, description } = nasData;
    
    const query = `
      UPDATE nas 
      SET shortname = ?, nasname = ?, secret = ?, type = ?, ports = ?, community = ?, description = ?
      WHERE id = ?
    `;
    
    const result = await db.query(query, [name, ip, secret, type, ports, community, description, id]);
    
    if (result.affectedRows === 0) {
      return null;
    }
    
    return await this.getById(id);
  }

  // Delete NAS
  static async delete(id) {
    const query = 'DELETE FROM nas WHERE id = ?';
    const result = await db.query(query, [id]);
    return result.affectedRows > 0;
  }

  // Check if NAS exists by shortname or IP
  static async exists(shortname, nasname, excludeId = null) {
    let query = 'SELECT id FROM nas WHERE (shortname = ? OR nasname = ?)';
    const params = [shortname, nasname];
    
    if (excludeId) {
      query += ' AND id != ?';
      params.push(excludeId);
    }
    
    const result = await db.query(query, params);
    return result.length > 0;
  }

  // Get NAS count
  static async getCount() {
    const query = 'SELECT COUNT(*) as count FROM nas';
    const result = await db.query(query);
    return result[0].count;
  }

  // Search NAS by name or IP
  static async search(searchTerm) {
    const query = `
      SELECT id, shortname, nasname, secret, type, ports, community, description 
      FROM nas 
      WHERE shortname LIKE ? OR nasname LIKE ? OR description LIKE ?
      ORDER BY shortname
    `;
    const searchPattern = `%${searchTerm}%`;
    return await db.query(query, [searchPattern, searchPattern, searchPattern]);
  }
}

module.exports = NasModel;