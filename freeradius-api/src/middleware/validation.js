const Joi = require('joi');

// Generic validation middleware
const validate = (schema, property = 'body') => {
  return (req, res, next) => {
    const { error, value } = schema.validate(req[property], {
      abortEarly: false,
      stripUnknown: true
    });
    
    if (error) {
      const errors = error.details.map(detail => ({
        field: detail.path.join('.'),
        message: detail.message
      }));
      
      return res.status(400).json({
        success: false,
        message: 'Validation error',
        errors
      });
    }
    
    req[property] = value;
    next();
  };
};

// NAS validation schemas
const nasSchemas = {
  create: Joi.object({
    name: Joi.string()
      .alphanum()
      .min(3)
      .max(30)
      .required()
      .messages({
        'string.alphanum': 'Name must contain only alphanumeric characters',
        'string.min': 'Name must be at least 3 characters long',
        'string.max': 'Name must not exceed 30 characters',
        'any.required': 'Name is required'
      }),
    ip: Joi.string()
      .ip({ version: ['ipv4', 'ipv6'] })
      .required()
      .messages({
        'string.ip': 'Must be a valid IP address',
        'any.required': 'IP address is required'
      }),
    secret: Joi.string()
      .min(8)
      .max(100)
      .required()
      .messages({
        'string.min': 'Secret must be at least 8 characters long',
        'string.max': 'Secret must not exceed 100 characters',
        'any.required': 'Secret is required'
      }),
    type: Joi.string()
      .valid('cisco', 'computone', 'livingston', 'juniper', 'max40xx', 'multitech', 'netserver', 'pathras', 'patton', 'portslave', 'tc', 'usrhiper', 'other')
      .default('other'),
    ports: Joi.number()
      .integer()
      .min(1)
      .max(65535)
      .default(1812),
    community: Joi.string()
      .max(50)
      .default(''),
    description: Joi.string()
      .max(200)
      .default('')
  }),
  
  update: Joi.object({
    name: Joi.string()
      .alphanum()
      .min(3)
      .max(30)
      .messages({
        'string.alphanum': 'Name must contain only alphanumeric characters',
        'string.min': 'Name must be at least 3 characters long',
        'string.max': 'Name must not exceed 30 characters'
      }),
    ip: Joi.string()
      .ip({ version: ['ipv4', 'ipv6'] })
      .messages({
        'string.ip': 'Must be a valid IP address'
      }),
    secret: Joi.string()
      .min(8)
      .max(100)
      .messages({
        'string.min': 'Secret must be at least 8 characters long',
        'string.max': 'Secret must not exceed 100 characters'
      }),
    type: Joi.string()
      .valid('cisco', 'computone', 'livingston', 'juniper', 'max40xx', 'multitech', 'netserver', 'pathras', 'patton', 'portslave', 'tc', 'usrhiper', 'other'),
    ports: Joi.number()
      .integer()
      .min(1)
      .max(65535),
    community: Joi.string()
      .max(50),
    description: Joi.string()
      .max(200)
  }).min(1)
};

// User validation schemas
const userSchemas = {
  create: Joi.object({
    user: Joi.string()
      .min(6)
      .required()
      .messages({
        'string.min': 'Username must be at least 6 characters long',
        'any.required': 'Username is required'
      }),
    password: Joi.string()
      .min(6)
      .required()
      .messages({
        'string.min': 'Password must be at least 6 characters long',
        'any.required': 'Password is required'
      }),
    profile: Joi.string()
  }),
  
  update: Joi.object({
    password: Joi.string()
      .min(6)
      .messages({
        'string.min': 'Password must be at least 6 characters long',
      }),
    profile: Joi.string()
  }).min(1)
};

// Authentication validation schemas
const authSchemas = {
  login: Joi.object({
    username: Joi.string()
      .required()
      .messages({
        'any.required': 'Username is required'
      }),
    password: Joi.string()
      .required()
      .messages({
        'any.required': 'Password is required'
      })
  })
};

// Query parameter validation schemas
const querySchemas = {
  pagination: Joi.object({
    page: Joi.number()
      .integer()
      .min(1)
      .default(1),
    limit: Joi.number()
      .integer()
      .min(1)
      .max(100)
      .default(10),
    search: Joi.string()
      .max(100)
      .optional()
  }),
  
  id: Joi.object({
    id: Joi.number()
      .integer()
      .min(1)
      .required()
      .messages({
        'number.base': 'ID must be a number',
        'number.integer': 'ID must be an integer',
        'number.min': 'ID must be greater than 0',
        'any.required': 'ID is required'
      })
  }),
  
  username: Joi.object({
    username: Joi.string()
      .min(6)
      .max(64)
      .required()
      .messages({
        'string.min': 'Username must be at least 6 characters long',
        'string.max': 'Username must not exceed 64 characters',
        'any.required': 'Username is required'
      })
  })
};

// Validation middleware functions
const validateNasCreate = validate(nasSchemas.create);
const validateNasUpdate = validate(nasSchemas.update);
const validateUserCreate = validate(userSchemas.create);
const validateUserUpdate = validate(userSchemas.update);
const validateLogin = validate(authSchemas.login);
const validatePagination = validate(querySchemas.pagination, 'query');
const validateId = validate(querySchemas.id, 'params');
const validateUsername = validate(querySchemas.username, 'params');

module.exports = {
  validate,
  validateNasCreate,
  validateNasUpdate,
  validateUserCreate,
  validateUserUpdate,
  validateLogin,
  validatePagination,
  validateId,
  validateUsername,
  schemas: {
    nas: nasSchemas,
    user: userSchemas,
    auth: authSchemas,
    query: querySchemas
  }
};