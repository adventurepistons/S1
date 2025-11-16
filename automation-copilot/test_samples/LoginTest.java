package com.example.tests;

import com.example.pages.LoginPage;
import org.openqa.selenium.WebDriver;
import org.openqa.selenium.chrome.ChromeDriver;
import org.testng.Assert;
import org.testng.annotations.*;

/**
 * Login Test Cases
 */
public class LoginTest {

    private WebDriver driver;
    private LoginPage loginPage;

    @BeforeMethod
    public void setUp() {
        driver = new ChromeDriver();
        driver.get("https://example.com/login");
        loginPage = new LoginPage(driver);
    }

    @AfterMethod
    public void tearDown() {
        if (driver != null) {
            driver.quit();
        }
    }

    @Test(description = "Verify successful login with valid credentials", priority = 1)
    public void testLoginWithValidCredentials() {
        // Arrange
        String username = "testuser@example.com";
        String password = "SecurePass123";

        // Act
        loginPage.login(username, password);

        // Assert
        Assert.assertEquals(driver.getCurrentUrl(), "https://example.com/dashboard", "User should be redirected to dashboard");
    }

    @Test(description = "Verify login fails with invalid password", priority = 2)
    public void testLoginWithInvalidPassword() {
        // Arrange
        String username = "testuser@example.com";
        String invalidPassword = "WrongPassword";

        // Act
        loginPage.login(username, invalidPassword);

        // Assert
        Assert.assertTrue(loginPage.isErrorMessageDisplayed(), "Error message should be displayed");
        Assert.assertEquals(loginPage.getErrorMessage(), "Invalid username or password", "Error message should match");
    }

    @Test(description = "Verify login fails with empty username", priority = 3)
    public void testLoginWithEmptyUsername() {
        // Arrange
        String emptyUsername = "";
        String password = "SecurePass123";

        // Act
        loginPage.enterUsername(emptyUsername);
        loginPage.enterPassword(password);
        loginPage.clickLoginButton();

        // Assert
        Assert.assertTrue(loginPage.isErrorMessageDisplayed(), "Error message should be displayed");
    }

    @Test(description = "Verify remember me functionality", priority = 4)
    public void testRememberMeCheckbox() {
        // Arrange
        String username = "testuser@example.com";
        String password = "SecurePass123";

        // Act
        loginPage.enterUsername(username);
        loginPage.enterPassword(password);
        loginPage.checkRememberMe();
        loginPage.clickLoginButton();

        // Assert
        Assert.assertEquals(driver.getCurrentUrl(), "https://example.com/dashboard", "User should be logged in");
    }

    @Test(description = "Verify error with invalid email format", priority = 5, enabled = true)
    public void testLoginWithInvalidEmailFormat() {
        // Arrange
        String invalidEmail = "notanemail";
        String password = "SecurePass123";

        // Act
        loginPage.login(invalidEmail, password);

        // Assert
        Assert.assertTrue(loginPage.isErrorMessageDisplayed(), "Error message should be displayed for invalid email");
    }
}
