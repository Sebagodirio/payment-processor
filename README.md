# Sistema de Pagos Orientado a Eventos (PoC en Go)
Este repositorio contiene una implementación de prueba de concepto (Proof of Concept) para un sistema de procesamiento de pagos, basado en una arquitectura de microservicios orientada a eventos y diseñada para ejecutarse en un entorno serverless (AWS Lambda).

El objetivo de este proyecto es demostrar la aplicación de principios de diseño de software robustos como la Arquitectura Hexagonal, el Patrón Saga para transacciones distribuidas y el manejo integral de errores en un sistema asíncrono.

## 1. Arquitectura del Sistema
La arquitectura general se basa en un conjunto de microservicios desacoplados que se comunican a través de un bus de eventos central. Cada servicio es dueño de su propio dominio y base de datos, garantizando una alta cohesión y bajo acoplamiento.

Para más detalles, consultar el documento de diseño de arquitectura completo. [Diseño completo](https://docs.google.com/document/d/1Q58RjFbj48WOTY-b6ZPQ5h_nuTJ1xJpIvmEAFc2aY98/edit?usp=sharing)

### Diagrama de Despliegue en AWS
![Diagrama de despliegue](img/image-1.png)

### Flujos de Pago
![Diagrama de secuencia - Happy path](img/image-2.png)
![Diagrama de secuencia - External provider error](img/image-3.png)
![Diagrama de secuencia - Fondos insuficientes](img/image-4.png)
![Diagrama de secuencia - Circuit breaker](img/image-5.png)
![Diagrama de secuencia - Pagos concurrentes](img/image-6.png)

## 2. Conceptos Clave
Arquitectura Orientada a Eventos (EDA): Los servicios reaccionan a eventos de negocio, lo que permite un desacoplamiento máximo y una alta escalabilidad.

Serverless: La implementación está pensada para AWS Lambda, eliminando la necesidad de gestionar servidores y permitiendo un escalado automático.

Arquitectura Hexagonal (Puertos y Adaptadores): La lógica de negocio (dominio) está completamente aislada de las tecnologías externas (frameworks, bases de datos, brokers de mensajes). La comunicación se realiza a través de puertos (interfaces) y adaptadores (implementaciones concretas), lo que facilita enormemente las pruebas y la mantenibilidad.

Patrón Saga (Coreografiada): Se utiliza para mantener la consistencia de los datos a través de múltiples servicios sin usar transacciones distribuidas. Si un paso falla, se ejecutan acciones de compensación.

## 3. Estructura del Proyecto (Monorepo)
El proyecto está organizado como un monorepo, donde cada servicio principal reside en su propio directorio.

/
├── cmd/
│   ├── payment/            # Entrypoint para la Lambda de Pagos
│   │   └── main.go
│   └── wallet/             # Entrypoint para la Lambda de Billetera
│       └── main.go
├── internal/
│   ├── config/             # Carga de configuración
│   ├── domain/             # Lógica de negocio pura, entidades y puertos
│   │   ├── payment/
│   │   └── wallet/
│   ├── platform/           # Implementación de adaptadores (infraestructura)
│   │   ├── bus/            # Adaptador para el bus de eventos (mock)
│   │   └── repository/     # Adaptador para la base de datos (mock)
│   └── services/           # Lógica de aplicación que conecta dominio e infraestructura
│       ├── payment/
│       └── wallet/
├── pkg/                    # Código compartido (ej. estructuras de eventos)
│   └── events/
│       └── events.go
└── go.mod

## 4. Contratos de Eventos
Los eventos son el contrato principal entre los servicios. Se definen con la siguiente estructura:

# Eventos de la Saga de Pago

| Nombre del Evento       | Servicio Publicador | Descripción                                                   |
|--------------------------|---------------------|---------------------------------------------------------------|
| **PaymentInit**          | Payment Service     | Inicia una nueva saga de pago.                               |
| **BalanceDebited**       | Wallet Service      | Notifica que el saldo ha sido debitado.                      |
| **InsufficientBalance**  | Wallet Service      | Notifica fallo por fondos insuficientes.                     |
| **ProviderPaymentSuccess** | Provider Gateway  | Confirma que la pasarela externa procesó el pago.            |
| **ProviderPaymentFailed**  | Provider Gateway  | Indica que la pasarela externa falló.                        |
| **ReembolsarUsuario**    | Payment Service     | Inicia la acción de compensación para devolver fondos.       |
| **PaymentCompleted**     | Payment Service     | Evento final que indica éxito en la saga.                    |
| **PaymentFailed**        | Payment Service     | Evento final que indica fallo en la saga.                    |



## 5. Alcance de la Implementación
Esta es una prueba de concepto y no una implementación lista para producción.

Servicios Implementados: payment-service, wallet-service (parcialmente).

Infraestructura: El bus de eventos y las bases de datos están simulados en memoria. No se realizan llamadas reales a servicios de AWS.

Observabilidad: La observabilidad se demuestra a través de logs estructurados en formato JSON que incluyen un correlationId para permitir el seguimiento de una transacción a través de los diferentes servicios.