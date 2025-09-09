const NasModel = require('../models/NasModel');

class NasController {
  // Get all NAS entries
  static async getAll(req, res) {
    try {
      const { search } = req.query;
      let nasEntries;
      
      if (search) {
        nasEntries = await NasModel.search(search);
      } else {
        nasEntries = await NasModel.getAll();
      }
      
      // Transform data to match API specification
      const transformedData = nasEntries.map(nas => ({
        id: nas.id,
        name: nas.shortname,
        ip: nas.nasname,
        secret: nas.secret,
        type: nas.type,
        ports: nas.ports,
        community: nas.community,
        description: nas.description
      }));
      
      res.json({
        success: true,
        message: 'NAS entries retrieved successfully',
        data: transformedData,
        count: transformedData.length
      });
    } catch (error) {
      console.error('Error getting NAS entries:', error);
      res.status(500).json({
        success: false,
        message: 'Internal server error',
        error: process.env.NODE_ENV === 'development' ? error.message : undefined
      });
    }
  }
  
  // Get NAS by ID
  static async getById(req, res) {
    try {
      const { id } = req.params;
      const nas = await NasModel.getById(id);
      
      if (!nas) {
        return res.status(404).json({
          success: false,
          message: 'NAS not found'
        });
      }
      
      // Transform data to match API specification
      const transformedData = {
        id: nas.id,
        name: nas.shortname,
        ip: nas.nasname,
        secret: nas.secret,
        type: nas.type,
        ports: nas.ports,
        community: nas.community,
        description: nas.description
      };
      
      res.json({
        success: true,
        message: 'NAS retrieved successfully',
        data: transformedData
      });
    } catch (error) {
      console.error('Error getting NAS by ID:', error);
      res.status(500).json({
        success: false,
        message: 'Internal server error',
        error: process.env.NODE_ENV === 'development' ? error.message : undefined
      });
    }
  }
  
  // Create new NAS
  static async create(req, res) {
    try {
      const { name, ip, secret, type, ports, community, description } = req.body;
      
      // Check if NAS already exists
      const exists = await NasModel.exists(name, ip);
      if (exists) {
        return res.status(409).json({
          success: false,
          message: 'NAS with this name or IP already exists'
        });
      }
      
      // Create new NAS
      const newNas = await NasModel.create({
        name,
        ip,
        secret,
        type,
        ports,
        community,
        description
      });
      
      // Transform data to match API specification
      const transformedData = {
        id: newNas.id,
        name: newNas.shortname,
        ip: newNas.nasname,
        secret: newNas.secret,
        type: newNas.type,
        ports: newNas.ports,
        community: newNas.community,
        description: newNas.description
      };
      
      res.status(201).json({
        success: true,
        message: 'NAS created successfully',
        data: transformedData
      });
    } catch (error) {
      console.error('Error creating NAS:', error);
      res.status(500).json({
        success: false,
        message: 'Internal server error',
        error: process.env.NODE_ENV === 'development' ? error.message : undefined
      });
    }
  }
  
  // Update NAS
  static async update(req, res) {
    try {
      const { id } = req.params;
      const updateData = req.body;
      
      // Check if NAS exists
      const existingNas = await NasModel.getById(id);
      if (!existingNas) {
        return res.status(404).json({
          success: false,
          message: 'NAS not found'
        });
      }
      
      // Check for conflicts with name or IP (excluding current NAS)
      if (updateData.name || updateData.ip) {
        const conflictName = updateData.name || existingNas.shortname;
        const conflictIp = updateData.ip || existingNas.nasname;
        
        const exists = await NasModel.exists(conflictName, conflictIp, id);
        if (exists) {
          return res.status(409).json({
            success: false,
            message: 'NAS with this name or IP already exists'
          });
        }
      }
      
      // Update NAS
      const updatedNas = await NasModel.update(id, updateData);
      
      if (!updatedNas) {
        return res.status(404).json({
          success: false,
          message: 'NAS not found or no changes made'
        });
      }
      
      // Transform data to match API specification
      const transformedData = {
        id: updatedNas.id,
        name: updatedNas.shortname,
        ip: updatedNas.nasname,
        secret: updatedNas.secret,
        type: updatedNas.type,
        ports: updatedNas.ports,
        community: updatedNas.community,
        description: updatedNas.description
      };
      
      res.json({
        success: true,
        message: 'NAS updated successfully',
        data: transformedData
      });
    } catch (error) {
      console.error('Error updating NAS:', error);
      res.status(500).json({
        success: false,
        message: 'Internal server error',
        error: process.env.NODE_ENV === 'development' ? error.message : undefined
      });
    }
  }
  
  // Delete NAS
  static async delete(req, res) {
    try {
      const { id } = req.params;
      
      // Check if NAS exists
      const existingNas = await NasModel.getById(id);
      if (!existingNas) {
        return res.status(404).json({
          success: false,
          message: 'NAS not found'
        });
      }
      
      // Delete NAS
      const deleted = await NasModel.delete(id);
      
      if (!deleted) {
        return res.status(404).json({
          success: false,
          message: 'NAS not found or already deleted'
        });
      }
      
      res.json({
        success: true,
        message: 'NAS deleted successfully'
      });
    } catch (error) {
      console.error('Error deleting NAS:', error);
      res.status(500).json({
        success: false,
        message: 'Internal server error',
        error: process.env.NODE_ENV === 'development' ? error.message : undefined
      });
    }
  }
  
  // Get NAS statistics
  static async getStats(req, res) {
    try {
      const count = await NasModel.getCount();
      
      res.json({
        success: true,
        message: 'NAS statistics retrieved successfully',
        data: {
          total_nas: count
        }
      });
    } catch (error) {
      console.error('Error getting NAS statistics:', error);
      res.status(500).json({
        success: false,
        message: 'Internal server error',
        error: process.env.NODE_ENV === 'development' ? error.message : undefined
      });
    }
  }
}

module.exports = NasController;