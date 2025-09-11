const db = require('../config/database');
const bcrypt = require('bcryptjs');

class UserModel {
  // Get all users with their profiles
  static async getAll() {
    const query = `
      SELECT DISTINCT 
        rc.id as id,
        rc.username as user,
        rc.value as password,
        rr.value as profile
      FROM radcheck rc
      LEFT JOIN radreply rr ON rc.username = rr.username AND rr.attribute = 'Mikrotik-Group'
      WHERE rc.attribute = 'Cleartext-Password'
      ORDER BY rc.username
    `;
    return await db.query(query);
  }

  // Get user by username
  static async getByUsername(username) {
    const query = `
      SELECT 
        rc.id as id,
        rc.username as user,
        rc.value as password,
        rr.value as profile
      FROM radcheck rc
      LEFT JOIN radreply rr ON rc.username = rr.username AND rr.attribute = 'Mikrotik-Group'
      WHERE rc.username = ? AND rc.attribute = 'Cleartext-Password'
    `;
    const result = await db.query(query, [username]);
    return result[0] || null;
  }

  // Get user by ID
  static async getById(id) {
    const query = `
      SELECT 
        rc.id as id,
        rc.username as user,
        rc.value as password,
        rr.value as profile
      FROM radcheck rc
      LEFT JOIN radreply rr ON rc.username = rr.username AND rr.attribute = 'Mikrotik-Group'
      WHERE rc.id = ? AND rc.attribute = 'Cleartext-Password'
    `;
    const result = await db.query(query, [id]);
    return result[0] || null;
  }

  // Create new user
  static async create(userData) {
    const { user, password, profile = 'PPP' } = userData;
    
    return await db.transaction(async (connection) => {
      // Insert into radcheck (username and password)
      const checkQuery = `
        INSERT INTO radcheck (username, attribute, op, value) 
        VALUES (?, 'Cleartext-Password', ':=', ?)
      `;
      const checkResult = await connection.execute(checkQuery, [user, password]);
      const userId = checkResult[0].insertId;
      
      // Insert into radreply (username and profile)
      const replyQuery = `
        INSERT INTO radreply (username, attribute, op, value) 
        VALUES (?, 'Mikrotik-Group', ':=', ?)
      `;
      await connection.execute(replyQuery, [user, profile]);
      
      return {
        id: userId,
        user,
        password,
        profile
      };
    });
  }

  // Update user
  static async update(username, userData) {
    const { password, profile } = userData;
    
    return await db.transaction(async (connection) => {
      let updated = false;
      
      // Update password if provided
      if (password) {
        const checkQuery = `
          UPDATE radcheck 
          SET value = ? 
          WHERE username = ? AND attribute = 'Cleartext-Password'
        `;
        const checkResult = await connection.execute(checkQuery, [password, username]);
        if (checkResult[0].affectedRows > 0) updated = true;
      }
      
      // Update profile if provided
      if (profile) {
        // Check if profile entry exists
        const existsQuery = `
          SELECT id FROM radreply 
          WHERE username = ? AND attribute = 'Mikrotik-Group'
        `;
        const existsResult = await connection.execute(existsQuery, [username]);
        
        if (existsResult[0].length > 0) {
          // Update existing profile
          const updateQuery = `
            UPDATE radreply 
            SET value = ? 
            WHERE username = ? AND attribute = 'Mikrotik-Group'
          `;
          const updateResult = await connection.execute(updateQuery, [profile, username]);
          if (updateResult[0].affectedRows > 0) updated = true;
        } else {
          // Insert new profile
          const insertQuery = `
            INSERT INTO radreply (username, attribute, op, value) 
            VALUES (?, 'Mikrotik-Group', ':=', ?)
          `;
          await connection.execute(insertQuery, [username, profile]);
          updated = true;
        }
      }
      
      if (!updated) {
        return null;
      }
      
      return await this.getByUsername(username);
    });
  }

  // Update user by ID
  static async updateById(id, userData) {
    const { password, profile } = userData;
    
    return await db.transaction(async (connection) => {
      let updated = false;
      
      // Get username from ID first
      const getUserQuery = `
        SELECT username FROM radcheck 
        WHERE id = ? AND attribute = 'Cleartext-Password'
      `;
      const userResult = await connection.execute(getUserQuery, [id]);
      if (userResult[0].length === 0) {
        return null;
      }
      const username = userResult[0][0].username;
      
      // Update password if provided
      if (password) {
        const checkQuery = `
          UPDATE radcheck 
          SET value = ? 
          WHERE id = ? AND attribute = 'Cleartext-Password'
        `;
        const checkResult = await connection.execute(checkQuery, [password, id]);
        if (checkResult[0].affectedRows > 0) updated = true;
      }
      
      // Update profile if provided
      if (profile) {
        // Check if profile entry exists
        const existsQuery = `
          SELECT id FROM radreply 
          WHERE username = ? AND attribute = 'Mikrotik-Group'
        `;
        const existsResult = await connection.execute(existsQuery, [username]);
        
        if (existsResult[0].length > 0) {
          // Update existing profile
          const updateQuery = `
            UPDATE radreply 
            SET value = ? 
            WHERE username = ? AND attribute = 'Mikrotik-Group'
          `;
          const updateResult = await connection.execute(updateQuery, [profile, username]);
          if (updateResult[0].affectedRows > 0) updated = true;
        } else {
          // Insert new profile
          const insertQuery = `
            INSERT INTO radreply (username, attribute, op, value) 
            VALUES (?, 'Mikrotik-Group', ':=', ?)
          `;
          await connection.execute(insertQuery, [username, profile]);
          updated = true;
        }
      }
      
      if (!updated) {
        return null;
      }
      
      return await this.getById(id);
    });
  }

  // Delete user
  static async delete(username) {
    return await db.transaction(async (connection) => {
      // Delete from radcheck
      const checkQuery = 'DELETE FROM radcheck WHERE username = ?';
      const checkResult = await connection.execute(checkQuery, [username]);
      
      // Delete from radreply
      const replyQuery = 'DELETE FROM radreply WHERE username = ?';
      await connection.execute(replyQuery, [username]);
      
      return checkResult[0].affectedRows > 0;
    });
  }

  // Check if user exists
  static async exists(username) {
    const query = 'SELECT username FROM radcheck WHERE username = ? AND attribute = "Cleartext-Password"';
    const result = await db.query(query, [username]);
    return result.length > 0;
  }

  // Get user count
  static async getCount() {
    const query = 'SELECT COUNT(DISTINCT username) as count FROM radcheck WHERE attribute = "Cleartext-Password"';
    const result = await db.query(query);
    return result[0].count;
  }

  // Search users
  static async search(searchTerm) {
    const query = `
      SELECT DISTINCT 
        rc.id as id,
        rc.username as user,
        rc.value as password,
        rr.value as profile
      FROM radcheck rc
      LEFT JOIN radreply rr ON rc.username = rr.username AND rr.attribute = 'Mikrotik-Group'
      WHERE rc.attribute = 'Cleartext-Password' AND rc.username LIKE ?
      ORDER BY rc.username
    `;
    const searchPattern = `%${searchTerm}%`;
    return await db.query(query, [searchPattern]);
  }

  // Authenticate user (for API access)
  static async authenticate(username, password) {
    const user = await this.getByUsername(username);
    if (!user) {
      return null;
    }
    
    // For FreeRADIUS, passwords are stored in plaintext in radcheck
    // In production, you might want to hash them
    if (user.password === password) {
      return user;
    }
    
    return null;
  }

  // Get user attributes (all radcheck entries for a user)
  static async getUserAttributes(username) {
    const query = 'SELECT attribute, op, value FROM radcheck WHERE username = ?';
    return await db.query(query, [username]);
  }

  // Get user reply attributes (all radreply entries for a user)
  static async getUserReplyAttributes(username) {
    const query = 'SELECT attribute, op, value FROM radreply WHERE username = ?';
    return await db.query(query, [username]);
  }

  // Add custom attribute to user
  static async addAttribute(username, attribute, op, value, table = 'radcheck') {
    const query = `INSERT INTO ${table} (username, attribute, op, value) VALUES (?, ?, ?, ?)`;
    const result = await db.query(query, [username, attribute, op, value]);
    return result.insertId;
  }

  // Remove custom attribute from user
  static async removeAttribute(username, attribute, table = 'radcheck') {
    const query = `DELETE FROM ${table} WHERE username = ? AND attribute = ?`;
    const result = await db.query(query, [username, attribute]);
    return result.affectedRows > 0;
  }
}

module.exports = UserModel;