package pages;


import org.openqa.selenium.By;
import org.openqa.selenium.WebDriver;

public class LoginPage {
    WebDriver driver;

    private final By emailField = By.id("Email");
    private final By passwordField = By.id("Password");
    private final By loginButton = By.cssSelector("input.login-button");
    private final By loginStatus = By.xpath("//a[@href=\"/customer/info\" and text()=\"nik@abc.com\"]");

    public LoginPage(WebDriver driver) {
        this.driver = driver;
    }

    public void enterEmail(String email) {
        driver.findElement(emailField).clear();
        driver.findElement(emailField).sendKeys(email);
    }

    public void enterPassword(String password) {
        driver.findElement(passwordField).clear();
        driver.findElement(passwordField).sendKeys(password);
    }

    public void clickLogin() {
        driver.findElement(loginButton).click();
    }

    public String getLoginStatusText() {
    	 System.out.println(driver.findElement(loginStatus).getText());
    
    
        return driver.findElement(loginStatus).getText();
    }
}
