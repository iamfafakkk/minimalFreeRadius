const UserModel = require('../models/UserModel');

class UserController {
  // Get all users
  static async getAll(req, res) {
    try {
      const { search } = req.query;
      let users;
      
      if (search) {
        users = await UserModel.search(search);
      } else {
        users = await UserModel.getAll();
      }
      
      res.json({
        success: true,
        message: 'Users retrieved successfully',
        data: users,
        count: users.length
      });
    } catch (error) {
      console.error('Error getting users:', error);
      res.status(500).json({
        success: false,
        message: 'Internal server error',
        error: process.env.NODE_ENV === 'development' ? error.message : undefined
      });
    }
  }
  
  // Get user by username
  static async getByUsername(req, res) {
    try {
      const { username } = req.params;
      const user = await UserModel.getByUsername(username);
      
      if (!user) {
        return res.status(404).json({
          success: false,
          message: 'User not found'
        });
      }
      
      res.json({
        success: true,
        message: 'User retrieved successfully',
        data: user
      });
    } catch (error) {
      console.error('Error getting user by username:', error);
      res.status(500).json({
        success: false,
        message: 'Internal server error',
        error: process.env.NODE_ENV === 'development' ? error.message : undefined
      });
    }
  }
  
  // Create new user
  static async create(req, res) {
    try {
      const { user, password, profile } = req.body;
      
      // Check if user already exists
      const exists = await UserModel.exists(user);
      if (exists) {
        return res.status(409).json({
          success: false,
          message: 'User already exists'
        });
      }
      
      // Create new user
      const newUser = await UserModel.create({
        user,
        password,
        profile
      });
      
      res.status(201).json({
        success: true,
        message: 'User created successfully',
        data: newUser
      });
    } catch (error) {
      console.error('Error creating user:', error);
      res.status(500).json({
        success: false,
        message: 'Internal server error',
        error: process.env.NODE_ENV === 'development' ? error.message : undefined
      });
    }
  }
  
  // Update user
  static async update(req, res) {
    try {
      const { username } = req.params;
      const updateData = req.body;
      
      // Check if user exists
      const existingUser = await UserModel.getByUsername(username);
      if (!existingUser) {
        return res.status(404).json({
          success: false,
          message: 'User not found'
        });
      }
      
      // Update user
      const updatedUser = await UserModel.update(username, updateData);
      
      if (!updatedUser) {
        return res.status(404).json({
          success: false,
          message: 'User not found or no changes made'
        });
      }
      
      res.json({
        success: true,
        message: 'User updated successfully',
        data: updatedUser
      });
    } catch (error) {
      console.error('Error updating user:', error);
      res.status(500).json({
        success: false,
        message: 'Internal server error',
        error: process.env.NODE_ENV === 'development' ? error.message : undefined
      });
    }
  }
  
  // Delete user
  static async delete(req, res) {
    try {
      const { username } = req.params;
      
      // Check if user exists
      const existingUser = await UserModel.getByUsername(username);
      if (!existingUser) {
        return res.status(404).json({
          success: false,
          message: 'User not found'
        });
      }
      
      // Delete user
      const deleted = await UserModel.delete(username);
      
      if (!deleted) {
        return res.status(404).json({
          success: false,
          message: 'User not found or already deleted'
        });
      }
      
      res.json({
        success: true,
        message: 'User deleted successfully'
      });
    } catch (error) {
      console.error('Error deleting user:', error);
      res.status(500).json({
        success: false,
        message: 'Internal server error',
        error: process.env.NODE_ENV === 'development' ? error.message : undefined
      });
    }
  }
  
  // Get user statistics
  static async getStats(req, res) {
    try {
      const count = await UserModel.getCount();
      
      res.json({
        success: true,
        message: 'User statistics retrieved successfully',
        data: {
          total_users: count
        }
      });
    } catch (error) {
      console.error('Error getting user statistics:', error);
      res.status(500).json({
        success: false,
        message: 'Internal server error',
        error: process.env.NODE_ENV === 'development' ? error.message : undefined
      });
    }
  }
  
  // Get user attributes (radcheck entries)
  static async getUserAttributes(req, res) {
    try {
      const { username } = req.params;
      
      // Check if user exists
      const user = await UserModel.getByUsername(username);
      if (!user) {
        return res.status(404).json({
          success: false,
          message: 'User not found'
        });
      }
      
      const attributes = await UserModel.getUserAttributes(username);
      
      res.json({
        success: true,
        message: 'User attributes retrieved successfully',
        data: {
          username,
          attributes
        }
      });
    } catch (error) {
      console.error('Error getting user attributes:', error);
      res.status(500).json({
        success: false,
        message: 'Internal server error',
        error: process.env.NODE_ENV === 'development' ? error.message : undefined
      });
    }
  }
  
  // Get user reply attributes (radreply entries)
  static async getUserReplyAttributes(req, res) {
    try {
      const { username } = req.params;
      
      // Check if user exists
      const user = await UserModel.getByUsername(username);
      if (!user) {
        return res.status(404).json({
          success: false,
          message: 'User not found'
        });
      }
      
      const replyAttributes = await UserModel.getUserReplyAttributes(username);
      
      res.json({
        success: true,
        message: 'User reply attributes retrieved successfully',
        data: {
          username,
          reply_attributes: replyAttributes
        }
      });
    } catch (error) {
      console.error('Error getting user reply attributes:', error);
      res.status(500).json({
        success: false,
        message: 'Internal server error',
        error: process.env.NODE_ENV === 'development' ? error.message : undefined
      });
    }
  }
  
  // Add custom attribute to user
  static async addAttribute(req, res) {
    try {
      const { username } = req.params;
      const { attribute, op, value, table = 'radcheck' } = req.body;
      
      // Check if user exists
      const user = await UserModel.getByUsername(username);
      if (!user) {
        return res.status(404).json({
          success: false,
          message: 'User not found'
        });
      }
      
      // Validate table parameter
      if (!['radcheck', 'radreply'].includes(table)) {
        return res.status(400).json({
          success: false,
          message: 'Table must be either "radcheck" or "radreply"'
        });
      }
      
      const attributeId = await UserModel.addAttribute(username, attribute, op, value, table);
      
      res.status(201).json({
        success: true,
        message: 'Attribute added successfully',
        data: {
          id: attributeId,
          username,
          attribute,
          op,
          value,
          table
        }
      });
    } catch (error) {
      console.error('Error adding user attribute:', error);
      res.status(500).json({
        success: false,
        message: 'Internal server error',
        error: process.env.NODE_ENV === 'development' ? error.message : undefined
      });
    }
  }
  
  // Remove custom attribute from user
  static async removeAttribute(req, res) {
    try {
      const { username } = req.params;
      const { attribute, table = 'radcheck' } = req.body;
      
      // Check if user exists
      const user = await UserModel.getByUsername(username);
      if (!user) {
        return res.status(404).json({
          success: false,
          message: 'User not found'
        });
      }
      
      // Validate table parameter
      if (!['radcheck', 'radreply'].includes(table)) {
        return res.status(400).json({
          success: false,
          message: 'Table must be either "radcheck" or "radreply"'
        });
      }
      
      const removed = await UserModel.removeAttribute(username, attribute, table);
      
      if (!removed) {
        return res.status(404).json({
          success: false,
          message: 'Attribute not found'
        });
      }
      
      res.json({
        success: true,
        message: 'Attribute removed successfully'
      });
    } catch (error) {
      console.error('Error removing user attribute:', error);
      res.status(500).json({
        success: false,
        message: 'Internal server error',
        error: process.env.NODE_ENV === 'development' ? error.message : undefined
      });
    }
  }
}

module.exports = UserController;