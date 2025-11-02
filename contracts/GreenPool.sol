// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/Ownable.sol";
import "./OXG.sol";

/**
 * @title GreenPool
 * @dev Fondo ambiental de Oxy•gen Economy
 * 
 * Recibe contribuciones automáticas del token OXG y distribuye fondos
 * a proyectos ambientales aprobados por la DAO
 */
contract GreenPool is Ownable {
    
    OXGToken public oxgToken;

    // Proyectos ambientales
    struct Project {
        address projectWallet;
        string name;
        string description;
        uint256 totalFunded;
        uint256 targetAmount;
        bool isActive;
        bool isApproved;
        uint256 createdAt;
    }

    mapping(uint256 => Project) public projects;
    uint256 public projectCount;
    
    // Distribuciones a proyectos
    struct Distribution {
        uint256 projectId;
        uint256 amount;
        uint256 timestamp;
        string description;
    }

    mapping(uint256 => Distribution[]) public projectDistributions;
    
    // Eventos
    event ProjectCreated(uint256 indexed projectId, address indexed wallet, string name);
    event ProjectApproved(uint256 indexed projectId);
    event ProjectFunded(uint256 indexed projectId, uint256 amount);
    event DistributionMade(uint256 indexed projectId, uint256 amount, string description);
    
    constructor(address _oxgToken) {
        require(_oxgToken != address(0), "OXG token address required");
        oxgToken = OXGToken(_oxgToken);
    }

    /**
     * @dev Crear un nuevo proyecto ambiental
     * @notice Solo el owner (o DAO) puede crear proyectos
     */
    function createProject(
        address _projectWallet,
        string memory _name,
        string memory _description,
        uint256 _targetAmount
    ) external onlyOwner returns (uint256) {
        require(_projectWallet != address(0), "Invalid project wallet");
        require(_targetAmount > 0, "Target amount must be greater than 0");

        projectCount++;
        projects[projectCount] = Project({
            projectWallet: _projectWallet,
            name: _name,
            description: _description,
            totalFunded: 0,
            targetAmount: _targetAmount,
            isActive: true,
            isApproved: false,
            createdAt: block.timestamp
        });

        emit ProjectCreated(projectCount, _projectWallet, _name);
        return projectCount;
    }

    /**
     * @dev Aprobar un proyecto (debería ser controlado por DAO)
     */
    function approveProject(uint256 _projectId) external onlyOwner {
        require(projects[_projectId].isActive, "Project does not exist");
        require(!projects[_projectId].isApproved, "Project already approved");
        
        projects[_projectId].isApproved = true;
        emit ProjectApproved(_projectId);
    }

    /**
     * @dev Distribuir fondos a un proyecto aprobado
     */
    function distributeToProject(
        uint256 _projectId,
        uint256 _amount,
        string memory _description
    ) external onlyOwner {
        Project storage project = projects[_projectId];
        
        require(project.isActive, "Project does not exist");
        require(project.isApproved, "Project must be approved");
        require(_amount > 0, "Amount must be greater than 0");
        
        uint256 balance = oxgToken.balanceOf(address(this));
        require(balance >= _amount, "Insufficient funds in GreenPool");

        // Transferir fondos al proyecto
        oxgToken.transfer(project.projectWallet, _amount);
        
        // Actualizar estadísticas
        project.totalFunded += _amount;
        
        // Registrar distribución
        projectDistributions[_projectId].push(Distribution({
            projectId: _projectId,
            amount: _amount,
            timestamp: block.timestamp,
            description: _description
        }));

        emit DistributionMade(_projectId, _amount, _description);
        emit ProjectFunded(_projectId, _amount);
    }

    /**
     * @dev Obtener balance del GreenPool
     */
    function getBalance() external view returns (uint256) {
        return oxgToken.balanceOf(address(this));
    }

    /**
     * @dev Obtener información de un proyecto
     */
    function getProject(uint256 _projectId) external view returns (Project memory) {
        return projects[_projectId];
    }

    /**
     * @dev Obtener distribuciones de un proyecto
     */
    function getProjectDistributions(uint256 _projectId) external view returns (Distribution[] memory) {
        return projectDistributions[_projectId];
    }

    /**
     * @dev Obtener todos los proyectos activos
     */
    function getActiveProjects() external view returns (Project[] memory) {
        Project[] memory activeProjects = new Project[](projectCount);
        uint256 count = 0;
        
        for (uint256 i = 1; i <= projectCount; i++) {
            if (projects[i].isActive) {
                activeProjects[count] = projects[i];
                count++;
            }
        }
        
        // Ajustar array al tamaño real
        Project[] memory result = new Project[](count);
        for (uint256 i = 0; i < count; i++) {
            result[i] = activeProjects[i];
        }
        
        return result;
    }
}

