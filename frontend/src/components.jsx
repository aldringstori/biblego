import React from 'react';

const baseStyles = {
    fontFamily: 'Arial, sans-serif',
    border: '1px solid #ccc',
    borderRadius: '4px',
    padding: '8px 12px',
    fontSize: '16px',
    margin: '5px 0',
};

export const Button = ({ children, onClick, type }) => (
    <button
        onClick={onClick}
        type={type}
        style={{
            ...baseStyles,
            backgroundColor: '#4CAF50',
            color: 'white',
            cursor: 'pointer',
            border: 'none',
            padding: '10px 20px',
        }}
    >
        {children}
    </button>
);

export const Select = ({ options, onChange, placeholder, style }) => (
    <select
        onChange={onChange}
        style={{...baseStyles, width: '100%', ...style}}
    >
        <option value="">{placeholder}</option>
        {options.map(option => (
            <option key={option.value} value={option.value}>{option.label}</option>
        ))}
    </select>
);

export const Input = ({ type, value, onChange, placeholder, style }) => (
    <input
        type={type}
        value={value}
        onChange={onChange}
        placeholder={placeholder}
        style={{...baseStyles, width: '100%', ...style}}
    />
);

export const Card = ({ children, style }) => (
    <div style={{
        ...baseStyles,
        boxShadow: '0 4px 8px 0 rgba(0,0,0,0.2)',
        transition: '0.3s',
        padding: '20px',
        ...style
    }}>
        {children}
    </div>
);

export const Title = ({ level, children }) => {
    const Tag = `h${level}`;
    return <Tag style={{ color: '#333', marginBottom: '20px' }}>{children}</Tag>;
};